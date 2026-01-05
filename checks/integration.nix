{
  pkgs,
  ...
}:
let
  webhooker = pkgs.callPackage ../packages/webhooker.nix {};
in
pkgs.testers.nixosTest {
  name = "webhooker-integration";
  globalTimeout = 120;

  nodes.server = {
    environment.systemPackages = [ webhooker pkgs.curl pkgs.jq ];
  };

  testScript = ''
    start_all()
    server.wait_for_unit("multi-user.target")

    # Create directories and start daemon
    server.succeed("mkdir -p /tmp/webhooker")
    server.succeed(
        "WEBHOOKER_ROUTES='secret123:github' "
        "WEBHOOKER_LOG_DIR=/tmp/webhooker "
        "WEBHOOKER_SOCKET=/tmp/webhooker.sock "
        "WEBHOOKER_PORT=8080 "
        "webhooker daemon &"
    )
    server.wait_for_open_port(8080)

    # Test 1: Persistent route
    server.succeed("curl -s -X POST -d '{\"repo\":\"test\"}' http://localhost:8080/secret123")
    server.succeed("grep -q 'github' /tmp/webhooker/webhook.log")

    # Test 2: Unknown route is ignored
    server.succeed("curl -s -X POST -d '{}' http://localhost:8080/unknown")
    server.succeed("test $(wc -l < /tmp/webhooker/webhook.log) -eq 1")

    # Test 3: Temporary route via IPC
    server.succeed(
        "WEBHOOKER_SOCKET=/tmp/webhooker.sock webhooker > /tmp/client.out 2>/tmp/client.err &"
    )
    server.sleep(1)

    # Get temp path from stderr
    temp_path = server.succeed("sed -n 's|.*\\(/tmp-[a-f0-9]*\\).*|\\1|p' /tmp/client.err").strip()

    # Send webhook to temp path
    server.succeed(f"curl -s -X POST -d '{{\"temp\":true}}' http://localhost:8080{temp_path}")
    server.sleep(1)

    # Verify client received it
    server.succeed("grep -q 'temp' /tmp/client.out")

    # Temp route should NOT be in persistent log
    server.succeed("! grep -q 'temp' /tmp/webhooker/webhook.log")
  '';
}
