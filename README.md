# Link Saver Desktop

A desktop client for [Link Saver](https://github.com/mikaelstaldal/linksaver), built with Go and GTK4.

## Features

- View a list of saved links and notes.
- Search through your items.
- Add new links by URL.
- Add new notes with a title and text.
- Edit existing items (title and description).
- Delete items.
- Settings dialog to configure API connection (Base URL, Username, Password).

## Requirements

- Go 1.25 or later.
- GTK4 libraries are installed on your system, see https://github.com/diamondburned/gotk4-examples/tree/master

## Build

To build the application, run:

```bash
go build -o linksaver-desktop .
```

## Usage

After building, you can run the application:

```bash
./linksaver-desktop
```

On the first run, go to **Settings** to configure your API endpoint and credentials. Settings are stored in your user configuration directory (e.g., `~/.config/linksaver/settings.json` on Linux).

## Development

### Dependencies

Dependencies are managed via Go modules. Run:

```bash
go mod tidy
```

### Testing

Run the standard Go tests:

```bash
go test ./...
```

Unit tests focus on logic separated from the UI.

## License

Copyright 2026 Mikael Ståldal.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
