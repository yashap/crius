import React from 'react';
import { render } from '@testing-library/react';
import App from './App';
import { Provider } from 'react-redux';
import { store } from '../store';

const WiredApp = () => (
  <React.StrictMode>
    <Provider store={store}>
      <App />
    </Provider>
  </React.StrictMode>
);

test('renders without errors', () => {
  const { getByText } = render(<WiredApp />);
  const linkElement = getByText(/Crius/i);
  expect(linkElement).toBeInTheDocument();
});
