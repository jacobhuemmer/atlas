class Atlas < Formula
  desc "Atlassian CLI for Jira, Confluence, Bitbucket, and JSM customer REST"
  homepage "https://github.com/masonhuemmer/atlas"
  url "https://github.com/masonhuemmer/atlas/archive/refs/tags/v1.3.1.tar.gz"
  sha256 "c6918b1d94293c4b1b47f56496feec4ca46eac9ac668cc4af848f5d6b679fd95"
  license "MIT"
  head "https://github.com/masonhuemmer/atlas.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/atlas"
  end

  test do
    assert_match "atlas", shell_output("#{bin}/atlas --help")
    assert_match "atlas://skill", shell_output("#{bin}/atlas mcp --help")
  end
end
