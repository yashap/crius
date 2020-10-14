import { List, Map } from 'immutable';
import * as rc from 'typed-rest-client';
import { Service, ServiceEndpoint, ServiceEndpointDependencies } from '../model';

/** The Crius HTTP API client */
class Crius {
  private readonly client: rc.RestClient;

  constructor(baseUrl: string) {
    this.client = new rc.RestClient('crius-ui', baseUrl);
  }

  async getAllServices() {
    const response = await this.client.get<ServiceDTO[]>('/services');
    const serviceDTOs: ServiceDTO[] = response.result ?? [];
    return List(serviceDTOs).map(serviceDTO => toDomainModel(serviceDTO));
  }
}

interface ServiceDTO {
  code: string;
  name: string;
  endpoints: ServiceEndpointDTO[];
}

interface ServiceEndpointDTO {
  code: string;
  name: string;
  serviceEndpointDependencies: ServiceEndpointDependenciesDTO;
}

type ServiceEndpointDependenciesDTO = Record<string, string[]>

const toDomainModel = (serviceDTO: ServiceDTO) => {
  return new Service({
    code: serviceDTO.code,
    name: serviceDTO.name,
    endpoints: List(serviceDTO.endpoints).map(endpointDTO => {
      const serviceEndpointDependencies: ServiceEndpointDependencies =
        Map(endpointDTO.serviceEndpointDependencies).map(depEndpoints => List(depEndpoints));
      return new ServiceEndpoint({
        code: endpointDTO.code,
        name: endpointDTO.name,
        serviceEndpointDependencies: serviceEndpointDependencies,
      })
    })
  });
}

export {
  Crius,
};
