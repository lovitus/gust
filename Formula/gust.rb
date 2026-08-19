class Gust < Formula
  desc "GOST fork with SSH relay fallback enhancements"
  homepage "https://github.com/lovitus/gust"
  version "3.2.11"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/lovitus/gust/releases/download/v3.2.11/gost-darwin-arm64-3.2.11.tar.gz"
      sha256 "bfce2ee6e104444cb660dd66134e83c45a12999663950e5a3fb52251b71213bb"
    end

    on_intel do
      url "https://github.com/lovitus/gust/releases/download/v3.2.11/gost-darwin-amd64-3.2.11.tar.gz"
      sha256 "7f61c251a925e280130df66fc302578242ae41f75a927c03419a95d6a9b0c5ca"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/lovitus/gust/releases/download/v3.2.11/gost-linux-arm64-3.2.11.tar.gz"
      sha256 "c4537b9bebc50136cd1e11035adf4e7182e34a1cf3bbf048113b60dd2d6fa4c0"
    end

    on_intel do
      url "https://github.com/lovitus/gust/releases/download/v3.2.11/gost-linux-amd64-3.2.11.tar.gz"
      sha256 "40f3d5d6eba3cce07733751dfe98c666da5ded4aced5b9e4243dfe00f2fa4baa"
    end
  end

  def install
    bin.install Dir["gost-*"].first => "gost"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/gost -V")
  end
end
