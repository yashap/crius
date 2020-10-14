import React from 'react';
import { CssBaseline, Grid } from '@material-ui/core';

import Header from './Header';
import Graph from './Graph';

const App = () => {
  return (
    <React.Fragment>
      <CssBaseline />
      <Grid>
        <Header />
        <Graph />
      </Grid>
    </React.Fragment>
  );
}

export default App;
