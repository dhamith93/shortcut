# Shortcut

Simple way to share **files** and **clipboard** with devices within a local network. 

## Usage

Run the ./shortcut executable. A browser window will be opened with the URL to connect from other devices. 

Default port is `:5500`, but it can be changed by editing the `config.json` file. Make sure to add the semicolon before the port when changing.

Files can be dragged and dropped / uploaded from the browser and can be uploaded/downloaded by anyone visiting the URL from the same network from any device. 

Device that runs the executable act as a centralized server, and the files will be uploaded to the `public/files/` dir in executable path. These files will be removed once the executable is stopped running or the next time it is running if executable crashed/force closed.
