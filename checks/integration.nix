{
  pkgs,
  ...
}:
let
  webhooker = pkgs.callPackage ../packages/webhooker.nix {};
in
pkgs.testers.nixosTest {
  name = "webhooker-integration";
  globalTimeout = 300;

  nodes.server = {
    environment.systemPackages = [ webhooker pkgs.curl ];
  };

  testScript = ''
    start_all()
    server.wait_for_unit("multi-user.target")

    server.succeed("webhooker daemon &")
    server.wait_for_open_port(8080, timeout=60)

    # Unknown route is ignored
    server.succeed("curl -s -X POST -d '{}' http://localhost:8080/unknown")

    # Ephemeral route via IPC
    server.succeed("webhooker > /tmp/client.out 2>/tmp/client.err &")
    server.sleep(1)

    temp_path = server.succeed("sed -n 's|.*\\(/tmp-[a-f0-9]*\\).*|\\1|p' /tmp/client.err").strip()

    server.succeed(f"curl -s -X POST -d '{{\"test\":true}}' http://localhost:8080{temp_path}")
    server.sleep(1)

    server.succeed("grep -q 'test' /tmp/client.out")

    server.execute("pkill webhooker")
  '';
}
