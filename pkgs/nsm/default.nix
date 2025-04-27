{ lib
, buildGoModule
, fetchFromGitHub
, nix
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

  vendorHash = null;  # Will be updated after first build

  ldflags = ["-s" "-w"];

  buildInputs = [ nix ];

  checkPhase = ''
    go test ./tests/unit/...
  '';

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
