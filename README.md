# go-edi

EDI Utils for performing common tasks on EDI messages.

### Common utilities:
* Marshal - convert EDI into JSON
* Unmarshal - convert JSON into EDI

### Install
To install the CLI locally, you can run the following command:
```bash
curl -s https://raw.githubusercontent.com/azarc-io/go-edi/refs/heads/main/cmd/edi/install | bash
```

To install a beta candidate:
```bash
curl -s https://raw.githubusercontent.com/azarc-io/go-edi/refs/heads/main/cmd/edi/install | bash -s -- --include-rc
```

### Usage

#### Marshal from JSON to EDI

To marshal your data from an input of JSON to EDI, you can run the following command:
```bash
cat my.json | edi marshal -s edi-schema_example_v1.json
```

#### Unmarshal from EDI to JSON

To unmarshal your data from an input of EDI to JSON, you can run the following command:
```bash
cat my.edi | edi unmarshal -s edi-schema_example_v1.json
```
