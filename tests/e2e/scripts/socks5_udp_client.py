#!/usr/bin/env python3
"""Exercise a SOCKS5 UDP association with an IP or domain destination."""

import argparse
import ipaddress
import os
import socket
import struct


def recv_exact(connection, size):
    data = b""
    while len(data) < size:
        chunk = connection.recv(size - len(data))
        if not chunk:
            raise RuntimeError("unexpected SOCKS5 control EOF")
        data += chunk
    return data


def encode_address(host, port):
    try:
        address = ipaddress.ip_address(host)
    except ValueError:
        encoded = host.encode("idna")
        if len(encoded) > 255:
            raise ValueError("domain name is too long")
        return b"\x03" + bytes([len(encoded)]) + encoded + struct.pack("!H", port)
    if address.version == 4:
        return b"\x01" + address.packed + struct.pack("!H", port)
    return b"\x04" + address.packed + struct.pack("!H", port)


def decode_address(data, offset):
    address_type = data[offset]
    offset += 1
    if address_type == 1:
        host = socket.inet_ntop(socket.AF_INET, data[offset : offset + 4])
        offset += 4
    elif address_type == 4:
        host = socket.inet_ntop(socket.AF_INET6, data[offset : offset + 16])
        offset += 16
    elif address_type == 3:
        length = data[offset]
        offset += 1
        host = data[offset : offset + length].decode("idna")
        offset += length
    else:
        raise RuntimeError("unknown SOCKS5 address type")
    port = struct.unpack("!H", data[offset : offset + 2])[0]
    return host, port, offset + 2


def dns_query(name, query_id):
    labels = b"".join(bytes([len(label)]) + label.encode("ascii") for label in name.rstrip(".").split("."))
    return struct.pack("!HHHHHH", query_id, 0x0100, 1, 0, 0, 0) + labels + b"\x00\x00\x01\x00\x01"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--proxy", required=True)
    parser.add_argument("--username", default="")
    parser.add_argument("--password", default="")
    parser.add_argument("--target", required=True)
    parser.add_argument("--verbose", action="store_true")
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("--payload")
    group.add_argument("--dns-name")
    args = parser.parse_args()

    proxy_host, proxy_port_text = args.proxy.rsplit(":", 1)
    target_host, target_port_text = args.target.rsplit(":", 1)
    proxy_port = int(proxy_port_text)
    target_port = int(target_port_text)

    control = socket.create_connection((proxy_host, proxy_port), timeout=5)
    control.settimeout(5)
    try:
        methods = b"\x00\x02" if args.username else b"\x00"
        control.sendall(b"\x05" + bytes([len(methods)]) + methods)
        version, method = recv_exact(control, 2)
        if version != 5 or method == 0xFF:
            raise RuntimeError("SOCKS5 method negotiation failed")
        if method == 2:
            username = args.username.encode("utf-8")
            password = args.password.encode("utf-8")
            control.sendall(b"\x01" + bytes([len(username)]) + username + bytes([len(password)]) + password)
            if recv_exact(control, 2) != b"\x01\x00":
                raise RuntimeError("SOCKS5 authentication failed")
        elif method != 0:
            raise RuntimeError("unexpected SOCKS5 authentication method")

        control.sendall(b"\x05\x03\x00" + encode_address("0.0.0.0", 0))
        header = recv_exact(control, 4)
        if header[:2] != b"\x05\x00":
            raise RuntimeError("SOCKS5 UDP ASSOCIATE failed")
        address_type = header[3]
        if address_type == 1:
            rest = recv_exact(control, 6)
        elif address_type == 4:
            rest = recv_exact(control, 18)
        elif address_type == 3:
            length = recv_exact(control, 1)
            rest = length + recv_exact(control, length[0] + 2)
        else:
            raise RuntimeError("invalid SOCKS5 relay address")
        relay_host, relay_port, _ = decode_address(bytes([address_type]) + rest, 0)
        if relay_host in ("0.0.0.0", "::"):
            relay_host = proxy_host
        if args.verbose:
            print("SOCKS5 UDP relay: {0}:{1}".format(relay_host, relay_port))

        query_id = struct.unpack("!H", os.urandom(2))[0]
        payload = args.payload.encode("utf-8") if args.payload is not None else dns_query(args.dns_name, query_id)
        packet = b"\x00\x00\x00" + encode_address(target_host, target_port) + payload
        udp = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        udp.settimeout(5)
        try:
            udp.sendto(packet, (relay_host, relay_port))
            response, _ = udp.recvfrom(65535)
        finally:
            udp.close()
        if response[:3] != b"\x00\x00\x00":
            raise RuntimeError("invalid SOCKS5 UDP response")
        _, _, payload_offset = decode_address(response, 3)
        response_payload = response[payload_offset:]
        if args.payload is not None:
            if response_payload != payload:
                raise RuntimeError("UDP echo payload mismatch")
            print(response_payload.decode("utf-8"))
        else:
            if len(response_payload) < 12 or struct.unpack("!H", response_payload[:2])[0] != query_id:
                raise RuntimeError("DNS transaction ID mismatch")
            _, flags, _, answers, _, _ = struct.unpack("!HHHHHH", response_payload[:12])
            if not flags & 0x8000 or answers < 1:
                raise RuntimeError("DNS response has no answer")
            print("DNS PASS: {0} answer(s)".format(answers))
    finally:
        control.close()


if __name__ == "__main__":
    main()
