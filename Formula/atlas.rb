class Atlas < Formula
  desc "Atlassian CLI for Jira, Confluence, Bitbucket, and JSM customer REST"
  homepage "https://github.com/jacobhuemmer/atlas"
  url "https://github.com/jacobhuemmer/atlas/archive/refs/tags/v1.2.0.tar.gz"
  sha256 "81ea4a5e14b4a8fc6b782347510a54159f0c06da9c7ed31011f73824aac6a260"
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
