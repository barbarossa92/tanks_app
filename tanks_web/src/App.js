import React, { Component } from 'react';
import Map from "./components/Map";
import Login from "./components/Login";


class App extends Component {
  constructor() {
    super();
    this.state = {
      username: null
    }
    this.setUsername = username => {
      this.setState({username: username});
    }
  }


  render() {
    const {username} = this.state;
    return (
      username ? <Map username={username}/> : <Login setUsername={this.setUsername}/>
    )
  }
}

export default App;
