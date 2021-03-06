// This file is automatically generated. Do not modify it manually.

const manifest = JSON.parse(`
{
    "id": "com.imc.mattermost-plugin-pingboard",
    "name": "Pingboard",
    "description": "Adds information from Pingboard to user popover.",
    "version": "0.0.3",
    "min_server_version": "5.12.0",
    "server": {
        "executables": {
            "linux-amd64": "server/dist/plugin-linux-amd64",
            "darwin-amd64": "server/dist/plugin-darwin-amd64",
            "windows-amd64": "server/dist/plugin-windows-amd64.exe"
        },
        "executable": ""
    },
    "webapp": {
        "bundle_path": "webapp/dist/main.js"
    },
    "settings_schema": {
        "header": "",
        "footer": "Specify the client secret either in the config file (with key 'pingboardApiClientSecret'), or in environment var 'MM_PLUGIN_PINGBOARD_CLIENT_SECRET'.",
        "settings": [
            {
                "key": "pingboardApiClientID",
                "display_name": "Pingboard API client ID",
                "type": "text",
                "help_text": "",
                "placeholder": "",
                "default": null
            }
        ]
    }
}
`);

export default manifest;
export const id = manifest.id;
export const version = manifest.version;
