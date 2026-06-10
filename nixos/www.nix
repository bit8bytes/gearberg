{
  buildpkgs,
  version,
  pname ? "gearberg-www",
}:
buildpkgs.buildGoModule {
  pname = pname;
  version = version;
  src = ./..;

  vendorHash = "sha256-6CWy3rpmbsoiBNMUJhEr7mImeZjmSR55U40GvLYFklE=";

  subPackages = ["cmd/www"];

  ldflags = [
    "-s"
    "-w"
    "-X main.revision=${version}"
  ];
  tags = ["sqlite"];

  nativeBuildInputs = [buildpkgs.tailwindcss_4];

  preBuild = ''
    tailwindcss -i ./internal/assets/css/index.css -o ./internal/assets/dist/index.css --minify
  '';

  env = {
    CGO_ENABLED = "0";
  };

  doCheck = true;
  checkFlags = ["-short"];

  meta = with buildpkgs.lib; {
    description = "Landing page for gearberg.org.";
    homepage = "https://gearberg.org";
    maintainers = with maintainers; [
      tobiasgleiter
    ];
  };
}
