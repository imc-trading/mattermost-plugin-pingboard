import {Store, Action} from 'redux';

import {GlobalState} from '@mattermost/types/lib/store';

import manifest from '@/manifest';

import {PluginRegistry} from '@/types/mattermost-webapp';
import Reducer from "@/reducers";
import UserAttribute from "@/components/user_attribute";

export default class Plugin {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        registry.registerReducer(Reducer);
        registry.registerPopoverUserAttributesComponent(UserAttribute);
    }
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void
    }
}

window.registerPlugin(manifest.id, new Plugin());
