class Atlas < Formula
  desc "Atlassian CLI for Jira, Confluence, Bitbucket, and JSM customer REST"
  homepage "https://github.com/jacobhuemmer/atlas"
  url "https://github.com/jacobhuemmer/atlas/archive/refs/tags/v1.1.0.tar.gz"
  sha256 "be5e8c791112dd1b3ed22761119b9277eae08d09046ceed6198973f04f32e48b"
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
