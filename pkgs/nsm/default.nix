{ lib
, buildGoModule
, fetchFromGitHub
}:

buildGoModule rec {
  pname = "nsm";
  version = "1.1.6";

  src = fetchFromGitHub {
    owner = "mdaashir";
    repo = "NSM";
    rev = "v${version}";
    hash = lib.fakeHash;  # Will be updated after first build
  };

  vendorHash = null;

  ldflags = ["-s" "-w"];

  meta = with lib; {
    description = "NSM (Nix Shell Manager) - A tool to manage Nix development environments";
    homepage = "https://github.com/mdaashir/NSM";
    changelog = "https://github.com/mdaashir/NSM/releases/tag/v${version}";
    license = licenses.mit;
    maintainers = with maintainers; [ mdaashir ];
    mainProgram = "nsm";
    platforms = platforms.all;
  };
}
