class Nextvibe < Formula
  desc "Agent-native project navigation CLI"
  homepage "https://github.com/nextvibe/nextvibe"
  url "https://github.com/nextvibe/nextvibe/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_RELEASE_TARBALL_SHA256"
  license "MIT"
  head "https://github.com/nextvibe/nextvibe.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", "-trimpath", "-ldflags=-s -w", "-o", bin/"nextvibe", "./cmd/nextvibe"
  end

  test do
    assert_match "NextVibe", shell_output("#{bin}/nextvibe --help")
  end
end
