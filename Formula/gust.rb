class Gust < Formula
  desc "GOST fork with SSH relay fallback enhancements"
  homepage "https://github.com/lovitus/gust"
  version "3.2.8"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/lovitus/gust/releases/download/v3.2.8/gost-darwin-arm64-3.2.8.tar.gz"
      sha256 "918176ebc8fa7e1a57537c89d95d2187c44ed46741d11c8e5c51171b4fa31315"
    end

    on_intel do
      url "https://github.com/lovitus/gust/releases/download/v3.2.8/gost-darwin-amd64-3.2.8.tar.gz"
      sha256 "928a7e264b388a9f77ed02f6aeffd0814bb19ea016fec5cd6b337741b8057243"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/lovitus/gust/releases/download/v3.2.8/gost-linux-arm64-3.2.8.tar.gz"
      sha256 "76b4b5f54b60d3a97f0d938da9e4da3358e0b2550319ae12d33a2a7cb37cf3b1"
    end

    on_intel do
      url "https://github.com/lovitus/gust/releases/download/v3.2.8/gost-linux-amd64-3.2.8.tar.gz"
      sha256 "eb8257fb58ecbaf51f25e75185187bafbeca661631b383ef5aa1da8141dc836f"
    end
  end

  def install
    bin.install Dir["gost-*"].first => "gost"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/gost -V")
  end
end
