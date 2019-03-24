import {id as pluginId} from './manifest';

export default class Plugin {}

window.registerPlugin(pluginId, new Plugin());
