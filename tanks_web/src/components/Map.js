import React, { Component } from 'react';
import wall_icon from "../brick-wall.png";
import rocket_icon from "../fireworks.png";
import tank_icon from "../tank.png";
import tank_info from "../tank-info.png";
import viewer_info from "../eye-icon.png";


class Map extends Component {
    constructor(props) {
        super(props);
        this.state = {
          bothsCount: 0,
          rectSize: 50,
          users: [],
          mapObj: [],
          logMessages: [],
          rating: {"1": "-----", "2": "-----", "3": "-------"},
          created: false,
          username: null,
          tanksCount: 0,
          viewersCount: 0
        }
        this.createOrDelete = (e) => {
          e.preventDefault();
          this.setState({...this.state, created: !this.state.created});
          window.localStorage.setItem("created", !this.state.created);
          this.connection.send(JSON.stringify({message: e.target.value, username: this.props.username}));
        }
        this.handleKeyUp = (e) => {
          switch(e.code) {
            case "KeyK":
              if (e.repeat) return;
              return this.sendAction("fire");
            case "KeyW":
              return this.sendAction("up");
            case "KeyS":
              return this.sendAction("down");
            case "KeyA":
              return this.sendAction("left");
            case "KeyD":
              return this.sendAction("right");
            default:
              return
          }
        }
        this.logout = e => {
          if (this.state.created) {
            this.createOrDelete(e)
          }
          this.props.logout(e);
        }
      }

      componentWillMount() {
        let created = window.localStorage.getItem("created");
        if(created === "true") this.setState({...this.state, created: true});
      }

      componentWillUnmount() {
        this.connection.close();
      }

      componentDidMount() {
          window.addEventListener("keypress", this.handleKeyUp);
          this.connection = new WebSocket(`ws://localhost:8000/ws?username=${this.props.username}`);
          this.connection.onopen = () => {
            console.log("Connected!");
          }
          this.connection.onmessage = e => {
            console.log(e);
            let data = JSON.parse(e.data)
            if (data.hasOwnProperty("dead") && data.dead === true) {
              this.setState({...this.state, created: false});
            } else {
              let state = {mapObj: data.map, logMessages: data.log, 
                            viewersCount: data.viewers_count, tanksCount: data.tanks_count}
              if (data.hasOwnProperty("rating")) {
                state = {...state, rating: data.rating}
              }
            this.setState(state);
            }
          }
      }
      checkRect(val) {
        if (typeof(val) === "string" && val === "wall") {
          return wall_icon;
        } else if (typeof(val) === "object") {
          if (val.hasOwnProperty("tank")) {
            return rocket_icon;
          } else  if (val.hasOwnProperty("tankType")) {
            return tank_icon;
          } else {
            return null
          }
        } else {
          return null
        }
      }

      checkRoute(val) {
        if (typeof(val) === "object" && val.hasOwnProperty("tankType")) {
            if (val.route === "right") {
              return "90"
            } else if (val.route === "down") {
              return "180"
            } else if (val.route === "left") {
              return "270"
            } else {
              return "0"
            }
        } else {
          return "0"
      }
    }
    
      sendAction(action) {
        if (this.state.created && this.connection.readyState === this.connection.OPEN) {
        this.connection.send(JSON.stringify({message: action, username: this.props.username}));
      }
      }
      render() {
        const {rectSize, mapObj, created} = this.state;
        let rects = [];
        let mapHeight = mapObj.length - 1
        let mapWidth = mapObj[0] ? mapObj[0].length - 1 : 0
        for (var i = 0; i <= mapHeight; i++){
          for (var j = 0; j <= mapWidth; j++) {
            rects.push(<rect key={`${i}${j}`} width={rectSize} height={rectSize} x={j * rectSize} y={i * rectSize} fill="green" stroke="black"/>)
            if(mapObj[i][j] !== "null") {
            rects.push(<image key={`${i}-${j}`} xlinkHref={this.checkRect(mapObj[i][j])} x={j * rectSize} y={i * rectSize} width={rectSize}  transform={`rotate(${this.checkRoute(mapObj[i][j])} ${j * rectSize + (rectSize / 2)} ${i * rectSize + (rectSize / 2)})`} height={rectSize}/>)
            }
          }
        }
        const log = this.state.logMessages ? this.state.logMessages.map((m) => <p>{m}</p>) : null;
        return (
          <div onKeyPress={this.handleKeyUp}>
          <div className="map" style={{float: "left", width: "80%"}}>
          <svg style={{border:'2px solid green', width: `${rectSize * mapWidth+rectSize}px`, height: `${rectSize * mapHeight+rectSize}px`}}>
            {rects}
          </svg>
          <button type="button" onClick={this.createOrDelete} value={!created ? "create" : "delete"}>{!created ? "Зайти на карту" : "Выйти с карты"}</button>
          <button type="button" onClick={this.logout} value="delete">Выйти</button>
          </div>
          <div className="log" style={{height: "350px", width: "20%", overflowY: "scroll"}}>
            {log}
          </div>
          <div className="info">
            <p><span><img src={tank_info} style={{height: "50px", width: "50px"}}/> - {this.state.tanksCount ? this.state.tanksCount : 0}</span></p>
            <p><span><img src={viewer_info} style={{height: "50px", width: "50px"}}/> - {this.state.viewersCount ? this.state.viewersCount : 0}</span></p>
          </div>
          </div>
        );
      }
}

export default Map;