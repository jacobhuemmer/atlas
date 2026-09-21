class Atlas < Formula
  desc "Atlassian CLI for Sesami and Garda sites"
  homepage "https://github.com/masonhuemmer/atlas"
  license "MIT"
  head "https://github.com/masonhuemmer/atlas.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/atlas"
  end

  test do
    assert_match "atlas", shell_output("#{bin}/atlas --help")
  end
end
