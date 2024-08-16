import React, { Component } from 'react';
import { Header } from './Header';
import { Main } from './Main';
import './App.css';

class App extends Component {
    constructor(props) {
        super(props);
        this.state = {
            view: 'allPosts', // Initialize view state
            isLoggedIn: false, // Track login status
        };
    }

    handleViewChange = (newView) => {
        this.setState({ view: newView }); // Update view state
    };

    handleLogin = () => {
        this.setState({ isLoggedIn: true }); // Set login status to true after successful login
    };

    render() {
        return (
            <div className="App">
                <Header
                    onViewChange={this.handleViewChange}
                    isLoggedIn={this.state.isLoggedIn} // Pass login status to Header
                />
                <Main
                    view={this.state.view}
                    onViewChange={this.handleViewChange}
                    onLogin={this.handleLogin} // Pass login handler to Main
                />
            </div>
        );
    }
}

export default App;
