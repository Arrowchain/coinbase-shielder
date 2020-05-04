```
     ___           ___                       ___           ___           ___           ___           ___
    /\  \         /\  \          ___        /\__\         /\  \         /\  \         /\  \         /\  \
   /::\  \       /::\  \        /\  \      /::|  |       /::\  \       /::\  \       /::\  \       /::\  \
  /:/\:\  \     /:/\:\  \       \:\  \    /:|:|  |      /:/\:\  \     /:/\:\  \     /:/\ \  \     /:/\:\  \
 /:/  \:\  \   /:/  \:\  \      /::\__\  /:/|:|  |__   /::\~\:\__\   /::\~\:\  \   _\:\~\ \  \   /::\~\:\  \
/:/__/ \:\__\ /:/__/ \:\__\  __/:/\/__/ /:/ |:| /\__\ /:/\:\ \:|__| /:/\:\ \:\__\ /\ \:\ \ \__\ /:/\:\ \:\__\
\:\  \  \/__/ \:\  \ /:/  / /\/:/  /    \/__|:|/:/  / \:\~\:\/:/  / \/__\:\/:/  / \:\ \:\ \/__/ \:\~\:\ \/__/
 \:\  \        \:\  /:/  /  \::/__/         |:/:/  /   \:\ \::/  /       \::/  /   \:\ \:\__\    \:\ \:\__\
  \:\  \        \:\/:/  /    \:\__\         |::/  /     \:\/:/  /        /:/  /     \:\/:/  /     \:\ \/__/
   \:\__\        \::/  /      \/__/         /:/  /       \::/__/        /:/  /       \::/  /       \:\__\
    \/__/         \/__/                     \/__/         ~~            \/__/         \/__/         \/__/
     ___           ___                       ___           ___       ___           ___           ___
    /\  \         /\__\          ___        /\  \         /\__\     /\  \         /\  \         /\  \
   /::\  \       /:/  /         /\  \      /::\  \       /:/  /    /::\  \       /::\  \       /::\  \
  /:/\ \  \     /:/__/          \:\  \    /:/\:\  \     /:/  /    /:/\:\  \     /:/\:\  \     /:/\:\  \
 _\:\~\ \  \   /::\  \ ___      /::\__\  /::\~\:\  \   /:/  /    /:/  \:\__\   /::\~\:\  \   /::\~\:\  \
/\ \:\ \ \__\ /:/\:\  /\__\  __/:/\/__/ /:/\:\ \:\__\ /:/__/    /:/__/ \:|__| /:/\:\ \:\__\ /:/\:\ \:\__\
\:\ \:\ \/__/ \/__\:\/:/  / /\/:/  /    \:\~\:\ \/__/ \:\  \    \:\  \ /:/  / \:\~\:\ \/__/ \/_|::\/:/  /
 \:\ \:\__\        \::/  /  \::/__/      \:\ \:\__\    \:\  \    \:\  /:/  /   \:\ \:\__\      |:|::/  /
  \:\/:/  /        /:/  /    \:\__\       \:\ \/__/     \:\  \    \:\/:/  /     \:\ \/__/      |:|\/__/
   \::/  /        /:/  /      \/__/        \:\__\        \:\__\    \::/__/       \:\__\        |:|  |
    \/__/         \/__/                     \/__/         \/__/     ~~            \/__/         \|__|
         _        ___         ___
__   __ / |      / _ \       / _ \
\ \ / / | |     | | | |     | | | |
 \ V /  | |  _  | |_| |  _  | |_| |
  \_/   |_| (_)  \___/  (_)  \___/

by j4ys0n

-amount int
  minimum amount to shield (default 1)
-batch string
  max number of coinbase UTXOs to shield at once (default "50")
-clipath string
  directory containing arrow-cli (no trailing /)
  default: /Applications/quiver.app/Contents/MacOS
  current: current working directory
  or specify your on directory path (default "default")
-from string
  address that has unshielded coinbase UTXOs
-help
  print this message
-minconfs int
  minimum block confirmations before shielding more (default 2)
-rpcconnect string
  ip address to connect to (default "127.0.0.1")
-rpcpass string
  password to use for connection
-rpcport string
  port to connect to (default "6543")
-rpcuser string
  user to use for connection
-to string
  address to send the shielded output to
-txpoll int
  interval in seconds at which to check if shielded tx confirmed (default 15)
```

## example usage

`shielderd -from arF1iEdxaSSwFaQYhnFHoKRcQoP4G44SDtE -to as1cdje006fn0pk74tfu5u5z354kqpdpgcu47qgcc9a62wre6r9cl5g7347f04pcmt8u6sw23sngj6 -amount 100 -batch 20`

if you have `quiver` installed into your `/Applications` folder on your mac, `arrow-cli` will be found there automatically.

you can place a copy of `arrow-cli` in the same directory as `shielderd` and pass `-clipath current` to `shielderd` and it will find it there.

or you can specify your own path like `-clipath /this/is/my/path`. do not add `arrow-cli` and do not add a trailing `/`.

## development

install packages

`go get -u ./...`

run locally

`go run shielderd.go`

build

`go build shielderd.go`
