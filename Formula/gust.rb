class Gust < Formula
  desc "GOST fork with SSH relay fallback enhancements"
  homepage "https://github.com/lovitus/gust"
  version "3.2.12"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/lovitus/gust/releases/download/v3.2.12/gost-darwin-arm64-3.2.12.tar.gz"
      sha256 "afca628aadfbf6f26922c1cdc97949acaf93ab092738bac40ee7d94286438696"
    end

    on_intel do
      url "https://github.com/lovitus/gust/releases/download/v3.2.12/gost-darwin-amd64-3.2.12.tar.gz"
      sha256 "0035c5687c5b12f00948f363433dcb86d84f4bf750d3cb51edd030baa99f083a"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/lovitus/gust/releases/download/v3.2.12/gost-linux-arm64-3.2.12.tar.gz"
      sha256 "8aad692bec3f8bf49bab1427526d0970f37ef77c413d87df5bd155048babf413"
    end

    on_intel do
      url "https://github.com/lovitus/gust/releases/download/v3.2.12/gost-linux-amd64-3.2.12.tar.gz"
      sha256 "858498226a0d9c1d112d69b89102157a49e6cfc2caf6144326c48e9a91e4a3ae"
    end
  end

  def install
    bin.install Dir["gost-*"].first => "gost"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/gost -V")
  end
end
