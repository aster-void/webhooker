{pkgs, ...}:
pkgs.buildGoModule {
  pname = "webhooker";
  version = "0.1.0";
  src = ../.;
  subPackages = [ "cmd/webhooker" ];
  vendorHash = null;
  meta = {
    description = "Simple webhook receiver with IPC streaming";
    mainProgram = "webhooker";
  };
}
