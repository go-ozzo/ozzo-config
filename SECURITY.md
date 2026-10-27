# Security Policy

## Supported Versions

No version has been tagged yet; fixes land on `master`. After the first release, the latest release is supported.

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Report it privately with **Report a vulnerability** on the repository's
[Security tab](https://github.com/go-ozzo/ozzo-config/security), which opens a private advisory draft:
https://github.com/go-ozzo/ozzo-config/security/advisories/new

Please include:

- a description of the vulnerability and its impact;
- a minimal reproduction (the configuration data and the code that loads or applies it);
- the affected versions you checked.

### Response Timeline

- **Initial response**: within 72 hours
- **Fix and disclosure**: coordinated with the reporter; credit is given unless you prefer otherwise

Published advisories: https://github.com/go-ozzo/ozzo-config/security/advisories

## Security Considerations

ozzo-config turns configuration into Go values. Things to keep in mind:

1. **Configuration is trusted input** — `Config.Load` opens exactly the path it is given, with no sanitizing, and the
   file extension selects the parser. Never build a path from untrusted input, and do not load configuration files
   that untrusted users can write.
2. **Parser resource use** — JSON, YAML and TOML documents are parsed in memory, and the package sets no limit on
   file size, nesting depth or the size of the resulting data. The YAML parser limits alias expansion ("billion
   laughs" documents are rejected), but a large or deeply nested document still costs memory and time in proportion
   to its size. Do not load untrusted documents without a size limit of your own.
3. **Configure sets exported fields only** — `Config.Configure` uses reflection and assigns only exported struct
   fields; a field that is missing or cannot be set is an error. A non-empty interface is filled with an object
   created by a provider function that you passed to `Config.Register`, selected by the `type` element of the
   configuration; an unregistered type name is an error. The configuration therefore cannot make the package create
   a type that you did not register, but a provider you register is called whenever its name appears in the data.
   An empty interface (`interface{}`) receives the configuration value as it is. This describes JSON and TOML data.
   With YAML data, whose mappings have `interface{}` keys, `Configure` does not yet select registered types by
   `type`, and a mapping key that is not a string makes it panic; the fix is planned before v1.0.0.
