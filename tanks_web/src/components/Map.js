import React, { Component } from 'react';
import wall_icon from "../brick-wall.png";
import rocket_icon from "../fireworks.png";
import tank_icon from "../tank.png";


class Map extends Component {
    constructor(props) {
        super(props);
        this.state = {
          bothsCount: 0,
          rectSize: 50,
          users: [],
          mapObj: [],
          userCoords: [],
          created: false
        }
        this.create = (e) => {
          e.preventDefault();
          this.setState({...this.state, created: true});
          this.connection.send(JSON.stringify({message: e.target.value, username: this.props.username}));
        }
        this.handleKeyUp = (e) => {
          switch(e.code) {
            case "Space":
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
      }
      componentDidMount() {
          window.addEventListener("keypress", this.handleKeyUp);
          this.connection = new WebSocket('ws://localhost:8000/ws');
          this.connection.onopen = () => {
            console.log("Connected!");
          }
          this.connection.onmessage = e => {
            let data = JSON.parse(e.data)
            console.log(e)
            this.setState({mapObj: data})
          }
      }
      checkRect(val) {
        if (typeof(val) === "string" && val === "wall") {
          return wall_icon;
        } else if (typeof(val) === "object") {
          if (val.hasOwnProperty("tank")) {
            return rocket_icon;
          } else  if (val.hasOwnProperty("tankType")) {
            if (this.props.username === val.name && this.state.userCoords !== val.coords){
              this.setState({...this.state, userCoords: val.coords})
            };
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
        this.connection.send(JSON.stringify({message: action, username: this.props.username, coords: this.state.userCoords}));
      }
      render() {
        const {rectSize, mapObj, created} = this.state;
        let rects = [];
        let mapHeight = Object.keys(mapObj).length - 1
        let mapWidth = mapObj["0"] ? Object.keys(mapObj["0"]).length - 1 : 0
        for (var i = 0; i <= mapHeight; i++){
          for (var j = 0; j <= mapWidth; j++) {
            rects.push(<rect key={i.toString() + j.toString()} width={rectSize} height={rectSize} x={i * rectSize} y={j * rectSize} fill="green" stroke="black"/>)
            if(mapObj[i][j] !== "null") {
            rects.push(<image xlinkHref={this.checkRect(mapObj[i][j])} x={i * rectSize} y={j * rectSize} width={rectSize}  transform={`rotate(${this.checkRoute(mapObj[i][j])} ${i * rectSize + (rectSize / 2)} ${j * rectSize + (rectSize / 2)})`} height={rectSize}/>)
            }
          }
        }
        return (
          <div onKeyPress={this.handleKeyUp}>
          <svg style={{border:'2px solid green', width: `${rectSize * mapWidth+50}px`, height: `${rectSize * mapHeight+50}px`}}>
            {rects}
          </svg>
          <button type="button" onClick={this.create} value="barbarossa">Barbarossa</button>
          {!created && <button type="button" onClick={this.create} value="create">Зайти</button>}
          </div>
          
        );
      }
}

export default Map;