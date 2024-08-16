import React from 'react';
import logo from './logo.svg';
import { Button } from 'antd';

export class Header extends React.Component {
    render() {
        return (
            <header className="App-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '20px', backgroundColor: '#000000', color: '#fff' }}>
                <div style={{ display: 'flex', alignItems: 'center' }}>
                    <h1 className="App-title" style={{ margin: 0 }}>ZConnect</h1>
                </div>
                {this.props.isLoggedIn && ( // Only show buttons if user is logged in
                    <div>
                        <Button style={{ marginRight: '10px' }} onClick={() => this.props.onViewChange('createPost')}>Create a New Post</Button>
                        <Button style={{ marginRight: '10px' }} onClick={() => this.props.onViewChange('myPosts')}>My Posts</Button>
                        <Button onClick={() => this.props.onViewChange('allPosts')}>All Posts</Button>
                    </div>
                )}
            </header>
        );
    }
}
