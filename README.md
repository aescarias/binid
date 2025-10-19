# BinID

BinID (Binary Identifier) is a tool for identifying and extracting the contents of a file.

BinID identifies files by relying on binary definition files written in a declarative language called [BinDef](./docs/bindef.adoc). Each definition describes a particular file format's features (e.g. magic numbers or signatures, headers, etc) and is compared with the input file. If a match is found, BinID will display details of the matched format plus file metadata.

BinID can be useful for finding software that can read a particular file, performing data recovery, reverse engineering or parsing a file, and more.

## Installation

For Windows users, you can download a built copy of BinID from the [Releases](https://github.com/aescarias/binid/releases) page. For other platforms, you will have to download the [Go](https://go.dev/) runtime, clone the project using `git clone`, and run `go build -o binid .\cmd` in the root directory to get a BinID executable for your platform.

You must also download the definitions folder to be able to use BinID. For now, we recommend that you `git clone` the repository or download the repository as a ZIP and extract the `formats` folder from it. In the future, a ready-to-go compressed archive will be provided for convenience.

## Usage

BinID can be invoked by doing `binid [filename]` where `[filename]` is the path to the file to identify.

BinID will attempt to load definitions from the `formats` folder in the directory where the executable is located. The `formats` folder contains the binary definitions that will be used by BinID for identifying files.

If BinID is able to identify a format, it will print information such as the example below:

```plaintext
found 33 definition(s)

== match
name: Executable and Linkable Format
mime(s): application/x-executable
details: The Executable and Linkable Format (ELF) is a common standard file format used for executable files and shared libraries. It is the standard executable file format among Unix-likes.

== metadata
Class: 2
Endianness: 1
Version: 1
Target ABI: 0
Target ABI version: 0
```

For certain byte sequences, the information will be stripped after 256 characters. Specifying the `-a` option will print the entire byte sequence, though note that this can produce fairly large outputs.
