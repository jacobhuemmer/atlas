class Atlas < Formula
  desc "Atlassian CLI for Jira, Confluence, Bitbucket, and JSM customer REST"
  homepage "https://github.com/jacobhuemmer/atlas"
  url "https://github.com/jacobhuemmer/atlas/archive/refs/tags/v1.3.0.tar.gz"
  sha256 "e7c43d483b66188748c8cdd6d3f0ec6f929a07dee21d88d58ad9cd9b5dd46552"
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
