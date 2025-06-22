class Markdocify < Formula
  desc "Convert documentation websites into consolidated markdown files"
  homepage "https://github.com/vladkampov/markdocify"
  version "0.0.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/vladkampov/markdocify/releases/download/v0.0.1/markdocify_0.0.1_darwin_arm64.tar.gz"
      sha256 "345be0b7c6ed261ac741aa17db94932b4b044fc164415abfdb645c2fc325221e"
    else
      url "https://github.com/vladkampov/markdocify/releases/download/v0.0.1/markdocify_0.0.1_darwin_amd64.tar.gz"
      sha256 "539d1efcc7c7cda7315c24abf7f4fe6ad593568b16546c8acd61b2a24e00313a"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/vladkampov/markdocify/releases/download/v0.0.1/markdocify_0.0.1_linux_arm64.tar.gz"
      sha256 "f52264d7b51d662939964b44f2c54695f86d6fd57876cdf8f986b179f7e49b4f"
    else
      url "https://github.com/vladkampov/markdocify/releases/download/v0.0.1/markdocify_0.0.1_linux_amd64.tar.gz"
      sha256 "037e80ea19202f2375b5163cb81f31964a0a6f020e21c194032d0742f4a60ec4"
    end
  end

  def install
    bin.install "markdocify"
    
    # Install config examples
    pkgshare.install "configs" if File.exist?("configs")
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/markdocify --version")
    
    # Test help output
    help_output = shell_output("#{bin}/markdocify --help")
    assert_match "markdocify is a CLI tool", help_output
  end
end