class Gust < Formula
  desc "GOST fork with SSH relay fallback enhancements"
  homepage "https://github.com/lovitus/gust"
  version "3.2.14"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/lovitus/gust/releases/download/v3.2.14/gost-darwin-arm64-3.2.14.tar.gz"
      sha256 "0dc469150a0015b01a024995ace156ebd2dfc40ef2bcad1d292468b31b65e84a"
    end

    on_intel do
      url "https://github.com/lovitus/gust/releases/download/v3.2.14/gost-darwin-amd64-3.2.14.tar.gz"
      sha256 "0680e95f45ae29d2cde230ef95381c0b70145c0e4c63ba1aa5c56defe052a34d"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/lovitus/gust/releases/download/v3.2.14/gost-linux-arm64-3.2.14.tar.gz"
      sha256 "5fbcf0a0118c9efab975a8f23d50b1d63336b71437252152f034dd4d8f51911d"
    end

    on_intel do
      url "https://github.com/lovitus/gust/releases/download/v3.2.14/gost-linux-amd64-3.2.14.tar.gz"
      sha256 "6fab5d42425f1427b4fd9d0c62ce2877e6f689b6ac6de4799fd55211d4312e77"
    end
  end

  def install
    bin.install Dir["gost-*"].first => "gost"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/gost -V")
  end
end
