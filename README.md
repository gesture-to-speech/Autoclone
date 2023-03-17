# Autoclone

## Run autoclone

### Add SSH keys
You need to create an SSH key for each repository you want to clone. To generate the key run `ssh-keygen -t rsa -b 4096`,
specify the path inside `.ssh` repository of the key and then click enter. Click enter again to create a key without
a passphrase and click enter again to confirm. A random SSH key will be generated. Then you will need to add it to
the `.ssh/config` [file](https://phoenixnap.com/kb/ssh-config). Repeat this process for each repository
you want to clone. You will need to send these keys also on GitHub. To do that, go to the repository's page (repeat for all repositories),
click `Settings -> Deploy keys (under Security on the left)`, and click button `Add deploy key`. 
Set the `Title` to `Autoclone` (or something similar), set `Key` to the output of this command `cat ~/.ssh/id_rsa.pub`, 
and click `Add key`.

### Setup autoclone
Clone this repository,  add `config.json` by copying the `config-template.json` file
and filling it with correct data (add `/` at the end of folder path), `key` for repositories should be empty.
Now you can run autoclone by running `./Autoclone`.

## Development
To install all dependencies for development: `sudo bash install.sh`.

Add bin path: `go env -w GOBIN=/path/to/folder/Autoclone/bin`

Compile to executable: `go install Autoclone`
