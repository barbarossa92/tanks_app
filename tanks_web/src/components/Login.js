import React, { Component } from 'react';


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
            <div className="login">
                <form className="login-form" onSubmit={this.submitForm}>
                    <label>Введите свое имя</label>
                    <input name="username" type="text" className="username" onChange={this.changeUsername}/>
                    <button className="login-submit">Войти</button>
                </form>
            </div>
        )
    }
}

export default Login;