import React, { Component } from 'react';
import "../styles.css"


class Login extends Component {
    constructor(props) {
        super(props);
        this.state = {
            username: null
        }
        this.submitForm = ev => {
            ev.preventDefault();
            this.state.username ? props.setUsername(this.state.username) : null
        }
        this.changeUsername = ev => {
            this.setState({username: ev.target.value});
        }
    }
    render() {
        return(
        <div className="wrapper">
            <form className="form-signin" onSubmit={this.submitForm}>       
                <h2 className="form-signin-heading">Введите имя</h2>
                <input type="text" className="form-control" name="username" placeholder="Имя" required autoFocus="" onChange={this.changeUsername}/> 
                <br/>   
                <button className="login-submit" type="submit">Войти</button>   
            </form>
        </div>
        )
    }
}

export default Login;