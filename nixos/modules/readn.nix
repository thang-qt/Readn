{ config, lib, pkgs, ... }:
let
  cfg = config.services.readn;
in
with lib; {
  options.services.readn = {
    enable = mkEnableOption "ReadN RSS reader service";

    package = mkOption {
      type = types.package;
      default = pkgs.readn;
      description = "Package providing the readn binary (defaults to overlay pkgs.readn).";
    };

    address = mkOption {
      type = types.str;
      default = "127.0.0.1:7070";
      description = "Address for ReadN to listen on (passed as --addr).";
      example = "127.0.0.1:8890";
    };

    basePath = mkOption {
      type = types.str;
      default = "";
      description = "Base path for the service URL (passed as --base).";
      example = "/read";
    };

    dbFile = mkOption {
      type = types.path;
      default = "/var/lib/readn/storage.db";
      description = "Path to the SQLite database file (passed as --db).";
    };

    authFile = mkOption {
      type = with types; nullOr path;
      default = null;
      description = "Optional path to auth file with username:password (passed as --auth-file).";
    };

    auth = mkOption {
      type = with types; nullOr str;
      default = null;
      description = "Optional username:password literal (passed as --auth).";
    };

    certFile = mkOption {
      type = with types; nullOr path;
      default = null;
      description = "Optional TLS certificate file (passed as --cert-file).";
    };

    keyFile = mkOption {
      type = with types; nullOr path;
      default = null;
      description = "Optional TLS key file (passed as --key-file).";
    };

    openBrowser = mkOption {
      type = types.bool;
      default = false;
      description = "Whether to open the app in a browser on start (passed as --open).";
    };
  };

  config = mkIf cfg.enable {
    systemd.services.readn = {
      description = "ReadN RSS Reader";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      serviceConfig = let
        flags = builtins.concatStringsSep " " (
          [
            "--addr=${cfg.address}"
            "--db=${toString cfg.dbFile}"
          ]
          ++ lib.optional (cfg.basePath != "") "--base=${cfg.basePath}"
          ++ lib.optional (cfg.authFile != null) "--auth-file=${toString cfg.authFile}"
          ++ lib.optional (cfg.auth != null) "--auth=${cfg.auth}"
          ++ lib.optional (cfg.certFile != null) "--cert-file=${toString cfg.certFile}"
          ++ lib.optional (cfg.keyFile != null) "--key-file=${toString cfg.keyFile}"
          ++ lib.optional cfg.openBrowser "--open"
        );
      in {
        ExecStart = ''${cfg.package}/bin/readn ${flags}'';
        Restart = "on-failure";
        DynamicUser = true;
        StateDirectory = "readn"; # creates /var/lib/readn with proper perms
        WorkingDirectory = "/var/lib/readn";
        # Hardened defaults
        ProtectSystem = "strict";
        ProtectHome = true;
        PrivateTmp = true;
        NoNewPrivileges = true;
        LockPersonality = true;
        MemoryDenyWriteExecute = true;
        RestrictRealtime = true;
        SystemCallFilter = [ "@system-service" ];
      };
    };
  };
}
