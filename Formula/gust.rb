class Gust < Formula
  desc "GOST fork with SSH relay fallback enhancements"
  homepage "https://github.com/lovitus/gust"
  version "3.2.10"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/lovitus/gust/releases/download/v3.2.10/gost-darwin-arm64-3.2.10.tar.gz"
      sha256 "598755b3432ba09ceba985e90510daaeaa7ccb4f81fb6efcc064dd3264c95dca"
    end

    on_intel do
      url "https://github.com/lovitus/gust/releases/download/v3.2.10/gost-darwin-amd64-3.2.10.tar.gz"
      sha256 "552a5522b9448530d2218f8472fedb41889545c001524da130fb041ffab5a0a5"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/lovitus/gust/releases/download/v3.2.10/gost-linux-arm64-3.2.10.tar.gz"
      sha256 "f52bad369b1806e064fee4f2c1c6e8f8ee5cfb69b0dd3493394edbbf5f9bbe53"
    end

    on_intel do
      url "https://github.com/lovitus/gust/releases/download/v3.2.10/gost-linux-amd64-3.2.10.tar.gz"
      sha256 "ba37f45a94d4c7f2864523eb8e1960a30abb8de597b05eb51ad552f94b36a76b"
    end
  end

  def install
    bin.install Dir["gost-*"].first => "gost"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/gost -V")
  end
end
