{
  config,
  lib,
  pkgs,
  ...
}:
let
  cfg = config.services.webhooker;
  defaultPackage = import ../../packages/webhooker.nix { inherit pkgs; };
in
{
  options.services.webhooker = {
    enable = lib.mkEnableOption "webhooker webhook receiver";

    package = lib.mkOption {
      type = lib.types.package;
      default = defaultPackage;
      description = "The webhooker package to use.";
    };

    port = lib.mkOption {
      type = lib.types.port;
      default = 8080;
      description = "HTTP port to listen on.";
    };

    domain = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "https://example.com";
      description = "Public base URL for webhook endpoints.";
    };
  };

  config = lib.mkIf cfg.enable {
    environment.systemPackages = [ cfg.package ];

    systemd.services.webhooker = {
      description = "Webhooker webhook receiver";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      environment = {
        WEBHOOKER_PORT = toString cfg.port;
      } // lib.optionalAttrs (cfg.domain != null) {
        WEBHOOKER_DOMAIN = cfg.domain;
      };

      serviceConfig = {
        Type = "simple";
        ExecStart = "${lib.getExe cfg.package} daemon";
        Restart = "on-failure";
        DynamicUser = true;
        RuntimeDirectory = "webhooker";
        RuntimeDirectoryMode = "0755";
      };
    };
  };
}
