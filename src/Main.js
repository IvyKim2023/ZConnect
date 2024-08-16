import React from 'react';
import Login from './Login';
import Register from './Register';
import PostsDisplay from './PostsDisplay';
import MyFriends from './MyFriends';

export class Main extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            showLogin: true,
            isLoggedIn: false,
        };
    }

    toggleView = () => {
        this.setState({ showLogin: !this.state.showLogin });
    };

    LoginSuccess = () => {
        this.setState({ isLoggedIn: true });
        this.props.onLogin(); // Notify App component that login was successful
    };

    render() {
        if (!this.state.isLoggedIn) {
            return (
                <div className="main">
                    {this.state.showLogin ? (
                        <Login LoginSuccess={this.LoginSuccess} toggleView={this.toggleView} />
                    ) : (
                        <Register toggleView={this.toggleView} />
                    )}
                </div>
            );
        }
        return (
            <div className="main">
                <div className="sidebar">
                    <MyFriends />
                </div>
                <div className="content">
                    <PostsDisplay view={this.props.view} onViewChange={this.props.onViewChange} />
                </div>
            </div>
        );
    }
}
