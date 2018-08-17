import React, { Component } from 'react';
import Map from "./components/Map";
import Login from "./components/Login";
import "./styles.css";


class App extends Component {
  constructor() {
    super();
    this.state = {
      username: null
    }
    this.setUsername = username => {
      let usrName = username.replace("-", "") + "-" + this.generateRandomString()
      window.localStorage.setItem("usrName", usrName);
      this.setState({username: usrName})
    }
    this.logout = e => {
      e.preventDefault();
      window.localStorage.removeItem("usrName");
      this.setState({username: null});
    }
  }

  componentWillMount() {
    let username = window.localStorage.getItem("usrName");
    if (username) {
      this.setState({username: username});
    }
  }

  generateRandomString() {
    let str = [...Array(10)].map(i=>(~~(Math.random()*36)).toString(36)).join('');
    return str;
  }


  render() {
    let {username} = this.state;
    return (
      username ? <Map username={this.state.username} logout={this.logout}/> : <Login setUsername={this.setUsername}/>
    )
  }
}

export default App;
