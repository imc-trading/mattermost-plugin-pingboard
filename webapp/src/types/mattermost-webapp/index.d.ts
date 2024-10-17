import {Reducer} from 'redux';

export interface PluginRegistry {
    registerPostTypeComponent(typeName: string, component: React.ElementType)
    registerReducer(reducer: Reducer)
    registerPopoverUserAttributesComponent(component: React.ReactNode)
}
