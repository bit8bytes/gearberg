{
  buildpkgs,
  version,
  pname ? "gearberg-docs",
}:
buildpkgs.buildGoModule {
  pname = pname;
  version = version;
  src = ./..;

  vendorHash = "sha256-2euSX31xv24c5w9XcYJmKR2xM0PuegN6obGaNOr1g+E=";

  subPackages = ["cmd/docs"];

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
    description = "Documentation page for gearberg.org.";
    homepage = "https://gearberg.org";
    maintainers = with maintainers; [
      tobiasgleiter
    ];
  };
}
