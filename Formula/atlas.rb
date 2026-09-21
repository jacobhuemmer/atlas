class Atlas < Formula
  desc "Atlassian CLI for Jira, Confluence, Bitbucket, and JSM customer REST"
  homepage "https://github.com/jacobhuemmer/atlas"
  url "https://github.com/jacobhuemmer/atlas/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "3b8ed50b75203eb6c37747907e6302ca190d52112c874e803a387363df7066b2"
  license "MIT"
  head "https://github.com/jacobhuemmer/atlas.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/atlas"
  end

  test do
    assert_match "atlas", shell_output("#{bin}/atlas --help")
    assert_match "atlas://skill", shell_output("#{bin}/atlas mcp --help")
  end
end
