import { configureStore, createAction, createAsyncThunk, createSlice, getDefaultMiddleware } from "@reduxjs/toolkit";
import { List, Record } from 'immutable';
import cytoscape from 'cytoscape';
import { Service } from '../model';
import { Crius } from '../api/crius';

class RootState extends Record({
  services: List<Service>(),
  selected: { key: '', value: '' } // TODO: immutable? Or should this just be a service code, forget key/value?
}) {
  asGraph(): DenormalizedElements {
    const edges: List<cytoscape.ElementDefinition> = this.services.flatMap(service => {
      return service.endpoints.flatMap(endpoint => {
        return endpoint.asEdges(service.code);
      });
    });
    return {
      nodes: this.services.map(service => service.asNode()).toJS(),
      edges: edges.toJS(),
    };
  }
}

type DenormalizedElements = {
  nodes: cytoscape.ElementDefinition[];
  edges: cytoscape.ElementDefinition[];
};

const criusClient: Crius = new Crius('http://localhost:3001'); // TODO: better way to do this

const actions = {
  selectService: createAction<{ key: string, value: string }>('crius/selectService'),
  fetchAllServices: createAsyncThunk<List<Service>>(
    'crius/getAllServicesStatus',
    async (arg, thunkAPI) => criusClient.getAllServices()
  )
};

const initialState: RootState = new RootState({});

const rootSlice = createSlice({
  name: 'crius',
  initialState: initialState,
  reducers: {}, // We use extraReducers with a builder for better type safety
  extraReducers: builder => {
    builder
      .addCase(actions.selectService, (state, action) => {
        return state.set('selected', action.payload);
      })
      .addCase(actions.fetchAllServices.pending, (state, action) => {
        // TODO: actually put it in a loading state
        return state.set('services', List<Service>());
      })
      .addCase(actions.fetchAllServices.fulfilled, (state, action) => {
        return state.set('services', action.payload);
      })
      .addCase(actions.fetchAllServices.rejected, (state, action) => {
        // TODO: actually put it in a "failed to load" state
        return state.set('services', List<Service>());
      });
  }
});

const middleware = getDefaultMiddleware({
  serializableCheck: false // Redux Toolkit complains about using Immutable JS records as state, but they're fine
});

const store = configureStore({
  reducer: rootSlice.reducer,
  middleware
});

export {
  RootState,
  store,
  actions,
};
