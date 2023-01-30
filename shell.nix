{ pkgs ? import <nixpkgs> { } }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    delve
    go_1_20
    golint
    gopls
  ];

  hardeningDisable = [ "fortify" ];
}
