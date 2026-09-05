{
  buildpkgs,
  version,
  pname ? "gearberg-docs",
}:
buildpkgs.buildGoModule {
  pname = pname;
  version = version;
  src = ./..;

  vendorHash = "sha256-+Hd/61Ux1QSCM4rLM7SobM78YcyxM0s47CqgHQy8j6c=";

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
