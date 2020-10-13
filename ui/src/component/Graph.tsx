import CytoscapeComponent from 'react-cytoscapejs';
import React, { Component } from 'react';
import { connect } from 'react-redux'
import cytoscape from 'cytoscape';
// @ts-ignore
import coseBilkentLayout from 'cytoscape-cose-bilkent';
import { Box, CircularProgress } from '@material-ui/core';
import { actions, RootState } from '../store';

import './Graph.css';

cytoscape.use(coseBilkentLayout);

class Graph extends Component<Props> {
  private cy: cytoscape.Core | null = null

  render = () => {
    const { elements } = this.props;
    if (elements.nodes.length > 0) {
      return (
        <div className="Graph">
          <CytoscapeComponent
            elements={CytoscapeComponent.normalizeElements(elements)}
            stylesheet={stylesheet}
            autoungrabify={true}
            layout={{ name: 'cose-bilkent' }}
            cy={(cy) => { this.cy = cy }}
            // @ts-ignore
            wheelSensitivity={0.3}
          />
        </div>
      );
    } else {
      return (
        <Box display="flex" justifyContent="center" alignItems="center" className="Graph">
          <CircularProgress />
        </Box>
      );
    }
  }

  componentDidMount() {
    this.props.fetchAllServices()
  }

  componentDidUpdate(prevProps: Readonly<Props>, prevState: Readonly<{}>, snapshot?: any) {
    this.maybePanToSelected(prevProps);
  }

  private getNode = (key: string, value: string) => this.cy?.$(`node[${key} = "${value}"]`);

  private panTo = (target: cytoscape.CollectionArgument) => {
    this.cy?.animate({
      fit: {
        eles: target,
        padding: 250
      },
      duration: 700,
      easing: 'ease',
      queue: true
    })
  }

  private maybePanToSelected = (prevProps: Readonly<Props>) => {
    const { selected } = this.props;
    if (selected) {
      const { key, value } = selected;
      const selectedDidUpdate = key !== prevProps.selected?.key || value !== prevProps.selected?.value;
      const nonEmptyArgs = key.length > 0 && value.length > 0;
      if (selectedDidUpdate && nonEmptyArgs) {
        const startNode = this.getNode(selected.key, selected.value);
        if (startNode) {
          this.panTo(startNode);
        }
      }
    }
  }
}

const mapStateToProps = (state: RootState) => {
  return {
    elements: state.asGraph(),
    selected: state.selected,
  };
};

const mapDispatchToProps = {
  fetchAllServices: actions.fetchAllServices
};

type StateProps = ReturnType<typeof mapStateToProps>
type DispatchProps = typeof mapDispatchToProps
type Props = StateProps & DispatchProps

// @ts-ignore
export default connect(mapStateToProps, mapDispatchToProps)(Graph);

const nodeStyle: cytoscape.Css.Node = {
  label: 'data(name)',
  'font-size': '4px',
};

const edgeStyle: cytoscape.Css.Edge = {
  label: 'data(name)',
  'curve-style': 'bezier',
  'font-size': '3px',
  'width': '1px',
  'target-arrow-shape': 'triangle',
  'arrow-scale': 0.5,
};

const stylesheet: cytoscape.Stylesheet[] = [
  {
    selector: 'node',
    style: nodeStyle,
  },
  {
    selector: 'edge',
    style: edgeStyle,
  }
];
