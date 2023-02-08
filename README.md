# /.idea Folder Enumerator

A simple go script that automatically enumerates an exposed `/.idea` folder from JetBrains IntelliJ.

The script checks the `/workspace.xml`, `/misc.xml`, and `/modules.xml` files for the given host, aggregates the file paths found in each file, checks for availability of the file paths, and downloads any files that were available. All files will be saved off in the directory structure listed in the found filepaths under the host name parent directory.

Supports proxy and running multiple threads.

Based loosely on the python version by @lijiegie 

## Usage

```
Usage:
    -host string
        Url to target. Example: https://example.com (default "REQUIRED")
    -proxy string
        Proxy host and port. Example: http://127.0.0.1:8080 (default: "NOPROXY")
    -threads int
        Number of concurrent threads to run. Example: 100 (default 50)
```

## Example

```
>> ./idea_enumerate -host https://example.com -proxy http://127.0.0.1:8080 -threads 100
[!] Found 427 filepaths
[!] Valid filepaths:
Downloaded file workspace.xml with size 35510
Downloaded file modules.xml with size 264
Downloaded file misc.xml with size 174
Downloaded file backup.sh withy size 8827
```
