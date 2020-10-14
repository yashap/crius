import React from 'react';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import { TextField } from '@material-ui/core';
import { Autocomplete } from '@material-ui/lab';
import { Service } from '../model';
import { actions, RootState } from '../store';
import { connect } from "react-redux";

const Header = ({ services, selectService }: Props) => {

  // TODO: pull out the service search component as its own component
  return (
    <AppBar position="absolute" color="default">
      <Toolbar>
        <Typography variant="h6" color="inherit" style={{ paddingRight : '20px' }} noWrap>
          Crius
        </Typography>
        <Autocomplete
          id="service-search"
          options={services.toJS()}
          getOptionLabel={(service) => service.name}
          style={{ width: 300 }}
          renderInput={(params) =>
            <TextField {...params} label="Find a service..." variant="outlined" />
          }
          onChange={(event, value) => value ? selectService(value) : null }
        />
      </Toolbar>
    </AppBar>
  );
}

const mapStateToProps = (state: RootState) => {
  return {
    services: state.services,
  };
};

const mapDispatchToProps = {
  selectService: (service: Service) => (actions.selectService({ key: 'code', value: service.code }))
};

type StateProps = ReturnType<typeof mapStateToProps>
type DispatchProps = typeof mapDispatchToProps
type Props = StateProps & DispatchProps

export default connect(mapStateToProps, mapDispatchToProps)(Header);
