import { List, Record } from 'immutable';
import cytoscape from "cytoscape";
import { ServiceEndpoint } from "./ServiceEndpoint";

class Service extends Record({
  code: '',
  name: '',
  endpoints: List<ServiceEndpoint>()
}) {
  asNode(): cytoscape.ElementDefinition {
    const data: cytoscape.NodeDataDefinition = {
      id: this.code, // TODO: return id from BE
      code: this.code,
      name: this.name,
      endpoints: this.endpoints, // TODO: should these be nested nodes?
    };
    return { data };
  }
}

export {
  Service,
};
