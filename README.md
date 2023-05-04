# Package Manager R Package Rewriting Utils

This repository includes shared code used to read and rewrite R source and
binary package tarballs and ZIP archives without the overhead of temporary
files and multiple passes. It supports:

- Reading and return DESCRIPTION information
- Modifying the DESCRIPTION file
- Extracting the package README
- Calculating both the original and resulting tarball SHA256 hashes.

This library is used by Posit Package Manager to extract README/DESCRIPTION
data and rewrite packages internally for local and Git sources. It is also used
by the Posit Package Service when creating CRAN and Bioconductor package
snapshots.
