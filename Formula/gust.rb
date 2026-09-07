class Gust < Formula
  desc "GOST fork with SSH relay fallback enhancements"
  homepage "https://github.com/lovitus/gust"
  version "3.2.13"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/lovitus/gust/releases/download/v3.2.13/gost-darwin-arm64-3.2.13.tar.gz"
      sha256 "8e8eaf9583965393c9223a88782b89918d3a724e8816361ec46d4673b1880593"
    end

    on_intel do
      url "https://github.com/lovitus/gust/releases/download/v3.2.13/gost-darwin-amd64-3.2.13.tar.gz"
      sha256 "db7c9557e141535c031091f4f835054df8578c58ecd588e826f9bd915df20d70"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/lovitus/gust/releases/download/v3.2.13/gost-linux-arm64-3.2.13.tar.gz"
      sha256 "532d0d020e362d1d2be96cd70a34f47e8b4d77d48e87cce4e187289ecc3b2a2d"
    end

    on_intel do
      url "https://github.com/lovitus/gust/releases/download/v3.2.13/gost-linux-amd64-3.2.13.tar.gz"
      sha256 "759fa73d758861117bbc77e3c146dfbd760dea8b8b22db83d14a0fedf5657963"
    end
  end

  def install
    bin.install Dir["gost-*"].first => "gost"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/gost -V")
  end
end
