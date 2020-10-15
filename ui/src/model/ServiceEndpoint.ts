import { List, Map, Record } from "immutable";
import cytoscape from "cytoscape";

class ServiceEndpoint extends Record({
  code: '',
  name: '',
  serviceEndpointDependencies: Map() as ServiceEndpointDependencies
}) {
  // TODO: I should just return the id from the BE
  id(serviceCode: string, endpointCode?: string): string {
    const ec = endpointCode ?? this.code;
    return `${serviceCode}|${ec}`;
  }

  asNode(serviceCode: string): cytoscape.ElementDefinition {
    const data: cytoscape.NodeDataDefinition = {
      id: this.id(serviceCode), // TODO: I should just return the id from the BE
      code: this.code,
      name: this.name,
      serviceEndpointDependencies: this.serviceEndpointDependencies,
    };
    return { data };
  }

  asEdges(serviceCode: string): cytoscape.ElementDefinition[] {
    const serviceEndpointDependencies: [string, List<string>][] = Array.from(this.serviceEndpointDependencies.entries());
    return serviceEndpointDependencies
      .flatMap(([depServiceCode, depEndpointCodes]) => {
        return depEndpointCodes.map(depEndpointCode => {
          const thisEndpointId = this.id(serviceCode); // TODO: return from BE
          const depEndpointId = this.id(depServiceCode, depEndpointCode); // TODO: return from BE
          const edgeId = `${thisEndpointId}=>${depEndpointId}`;
          const data: cytoscape.EdgeDataDefinition = {
            id: edgeId,
            source: serviceCode, // TODO: link endpoints, not services
            target: depServiceCode, // TODO: link endpoints, not services
            name: depEndpointCode,
          };
          return { data } as cytoscape.ElementDefinition;
        }).toJS();
      });
  }
}

export type ServiceEndpointDependencies = Map<string, List<string>>

export {
  ServiceEndpoint,
};
