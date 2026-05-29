{
  buildpkgs,
  version,
  pname ? "gearberg-www",
}:
buildpkgs.buildGoModule {
  pname = pname;
  version = version;
  src = ./..;

  vendorHash = "sha256-/rLXKPc0el25tamhckO7oLvJPFiMe9v8ufbMOJ31h4s=";

  ldflags = [
    "-s"
    "-w"
    "-X main.revision=${version}"
  ];
  tags = ["sqlite"];

  nativeBuildInputs = [buildpkgs.tailwindcss_4];

  preBuild = ''
    tailwindcss -i ./assets/css/index.css -o ./assets/dist/index.css --minify
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
