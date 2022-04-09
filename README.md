## Building Application
This guide will step you through the build process for this application.

1. Follow the instructions on [https://go.dev/doc/install](https://go.dev/doc/install) to install a `Go` compiler on your computer.
2. Clone this repo.
    ```bash
    git clone https://github.com/alexwilkerson/ddstats-go.git
    ```
3. `cd` into the `cmd/client` folder of this repository.
    ```bash
    cd ddstats-go/cmd/client
    ```
4. Build the application using `Go`.
    ```bash
    go build
    ```
    - This will generate `client.exe`. You can run this file from anywhere as long as `ddstats-go/config.toml` is in the same directory.

## Updating the BaseAddress
This section is working as of DevilDaggers build `8351000`.

1. Open CheatEngine and select `dd.exe` from the process list.
2. In the middle of the CheatEngine window should be a button labeled `Memory View`. Click it.
    - If CheatEngine has since changed its layout, you will want to find and open the `Memory Viewer` window.
4. In the middle of the `Memory Viewer` window you will see `AllocationBase` and `Base` addresses.
    - Subtract `Base` from `AllocationBase` using a hexidecimal calculator. The result is your `baseOffset`
5. Open `devildaggers/devildaggers.go`.
6. Near the top of this file, update the `baseOffset` value to match the result from step 3.
    - Make sure to pad the value to the right of `0x` with 0's to 8-digits. For example, if our `baseOffset` is `24F000`, we would write `0x0024F000`.
7. Open a terminal to `cmd/client` and build the project using `go`.
    ```bash
    cd cmd/client
    go build
    ```
8. Finally, copy the resulting `client.exe` along with the `config.toml` in the root of this repository into your preferred location. You may now run `client.exe`.