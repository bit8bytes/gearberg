# This is an example for building a Go app using Nix.
# Pass the buildpkgs, the version amd the name to the flake.
{
  buildpkgs,
  version,
  pname ? "gearberg",
}:
buildpkgs.buildGoModule {
  pname = pname;
  version = version;
  # Because the go.mod file is in the parent folder, we need to navigate
  # one folder up.
  src = ./..;

  vendorHash = "sha256-5ABKSj01op4mlyX5KFRkxeen2hz4NlHh3pTuZ/xqvvw=";

  subPackages = ["cmd/web"];

  # Buildflags and tags, the same as typical go build .
  ldflags = [
    "-s"
    "-w"
    "-X main.revision=${version}"
  ];
  tags = ["sqlite"];

  nativeBuildInputs = [buildpkgs.tailwindcss_4 buildpkgs.sqlc buildpkgs.ogen];

  preBuild = ''
    go generate ./api/
    sqlc generate -f sqlc.sqlite.yml
    tailwindcss -i ./assets/css/index.css -o ./assets/dist/index.css --minify
  '';

  env = {
    CGO_ENABLED = "0";
  };

  # After the build, we rename the binary to allow
  # users to call it e.g. nix run github:bit8bytes/gearberg
  postInstall = ''
    mv $out/bin/web $out/bin/gearberg
  '';

  doCheck = true;
  checkFlags = ["-short"];

  # Adding some metadata to the package.
  meta = with buildpkgs.lib; {
    description = "Self-hostable inventory and rental management.";
    homepage = "https://gearberg.org";
    maintainers = with maintainers; [
      tobiasgleiter
    ];
  };
}
