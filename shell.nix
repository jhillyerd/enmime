{ pkgs ? import <nixpkgs> { } }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    delve
    go
    golangci-lint
    gopls
  ];

  GOTOOLCHAIN = "go1.25.5";
  hardeningDisable = [ "fortify" ];
}
