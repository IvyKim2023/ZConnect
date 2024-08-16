import React, { useState, useEffect } from 'react';
import { List, Button, Input } from 'antd';
import $ from 'jquery';
import { API_ROOT } from './constant';
import PostsDisplay, { fetchOthersPosts } from './PostsDisplay';



const FriendsList = () => {
    const [friends, setFriends] = useState([]);
    const [newFriend, setNewFriend] = useState('');
    const [selectedFriend, setSelectedFriend] = useState(null);
    const [friendPosts, setFriendPosts] = useState([]);

    const fetchFriends = () => {
        const token = localStorage.getItem('token'); // Retrieve token from localStorage
        $.ajax({
            url: `${API_ROOT}/getfriends`, // API endpoint to get friends
            method: 'GET', // HTTP method
            headers: {
                'Authorization': `bearer ${token}`,
                'Content-Type': 'application/json', // Add the token to the Authorization header
            },
            success: (response) => {
                setFriends(response); // Update friends state with the response data
            },
            error: (error) => {
                console.error('Get Friends failed: ', error);
            },
        });
    };

    useEffect(() => {
        fetchFriends(); // Fetch friends when the component mounts
    }, []);// Empty dependency array ensures this runs only once after initial render

    const handleAdd = () => {
        const newFriendData = { friend_username: newFriend };
        const token = localStorage.getItem('token');
        $.ajax({
            url: `${API_ROOT}/addfriend`,
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
            },
            data: JSON.stringify(newFriendData),
            dataType: 'text',
            success: (response) => {
                console.log('Add friend response: ', response);
                fetchFriends(); // Call fetchFriends after the request completes
                setNewFriend(''); // Clear the input field
            },
            error: () => {
                console.log('Add friend failed');
            }
        });
    };

    const handleDelete = (friend) => {
        const token = localStorage.getItem('token');
        $.ajax({
            url: `${API_ROOT}/deletefriend`,
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
            },
            data: JSON.stringify({ friend_username: friend }),
            dataType: 'text',
            success: (response) => {
                console.log('Delete friend response: ', response);
                fetchFriends(); // Call fetchFriends after the request completes
            },
            error: () => {
                console.log('Delete friend failed');
            }
        });
    };




    return (
        <div className="friends-list-container">
            <header className="friends-list-header">My Friends</header>
            <div className="friends-list">
                <List
                    itemLayout="horizontal"
                    dataSource={friends}
                    renderItem={(friend) => (
                        <List.Item
                            actions={[<Button onClick={() => handleDelete(friend)}>Delete</Button>]}
                        >
                            <p >
                                {friend}
                            </p>
                        </List.Item>
                    )}
                />
                <Input
                    placeholder="Enter friend's name"
                    value={newFriend}
                    onChange={(e) => setNewFriend(e.target.value)}
                    style={{width: '200px', marginRight: '8px', marginBottom: '8px'}}
                />
                <Button type="primary" onClick={handleAdd}>
                    Add Friend
                </Button>
            </div>
        </div>
    );
};

export default FriendsList;
