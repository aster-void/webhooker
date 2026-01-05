{
  config,
  lib,
  pkgs,
  ...
}:
let
  cfg = config.services.webhooker;
  routesStr = lib.concatStringsSep "," (
    lib.mapAttrsToList (name: secret: "${secret}:${name}") cfg.routes
  );
in
{
  options.services.webhooker = {
    enable = lib.mkEnableOption "webhooker webhook receiver";

    package = lib.mkPackageOption pkgs "webhooker" { };

    port = lib.mkOption {
      type = lib.types.port;
      default = 8080;
      description = "HTTP port to listen on.";
    };

    dataDir = lib.mkOption {
      type = lib.types.path;
      default = "/var/lib/webhooker";
      description = "Base directory for data and logs.";
    };

    routes = lib.mkOption {
      type = lib.types.attrsOf lib.types.str;
      default = { };
      example = {
        github = "secret123";
        gitlab = "secret456";
      };
      description = ''
        Persistent routes as name-secret pairs.
        Format: { name = "secret"; }
      '';
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.webhooker = {
      description = "Webhooker webhook receiver";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      environment = {
        WEBHOOKER_PORT = toString cfg.port;
        WEBHOOKER_DATA_DIR = cfg.dataDir;
        WEBHOOKER_ROUTES = routesStr;
      };

      serviceConfig = {
        Type = "simple";
        ExecStart = lib.getExe cfg.package;
        Restart = "on-failure";
        DynamicUser = true;
        StateDirectory = "webhooker";
        RuntimeDirectory = "webhooker";
      };
    };
  };
}
