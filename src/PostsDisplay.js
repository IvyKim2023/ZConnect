import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Button, Alert } from 'antd';
import $ from "jquery";
import { API_ROOT } from "./constant";
import UploadButton from "./UploadButton";

// Create a New Post Component
const CreatePost = ({ onPostCreated }) => {
    const [message, setMessage] = useState('');
    const [image, setImage] = useState(null);
    const [isSuccessful, setIsSuccessful] = useState(false); // State to track success

    const handleSubmit = () => {
        const token = localStorage.getItem('token');
        const formData = new FormData();
        formData.append('message', message);
        if (image) {
            formData.append('image', image);
        }

        $.ajax({
            url: `${API_ROOT}/post`,
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
            },
            data: formData,
            processData: false, // Prevent jQuery from processing the data
            contentType: false, // Prevent jQuery from setting content type
            success: () => {
                setIsSuccessful(true); // Set success state to true
                onPostCreated(); // Switch back to post display after creation
            },
            error: (error) => {
                console.error('Create post failed:', error);
                setIsSuccessful(true); // Set success state to true
                onPostCreated(); // Switch back to post display after creation
            },
            complete: (xhr, status) => {
                if (status === 'nocontent' || xhr.status === 204) {
                    setIsSuccessful(true); // Handle cases where the server returns no content
                    onPostCreated();
                }
            }
        });
    };

    return (
        <div style={{ padding: '20px' }}>
            <h2>Create a New Post</h2>
            {isSuccessful && (
                <Alert message="Create Successful" type="success" showIcon style={{ marginBottom: '10px' }} />
            )}
            <input
                type="text"
                placeholder="Enter your message"
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                style={{ width: '100%', marginBottom: '10px', padding: '8px' }}
            />
            <UploadButton onFileSelect={setImage} />
            <Button type="primary" onClick={handleSubmit}>
                Submit
            </Button>
        </div>
    );
};

// Posts Display Component
const PostsDisplay = ({ view, selectedFriend }) => {
    const [posts, setPosts] = useState([]); // Initialize as an empty array

    useEffect(() => {
        if (view === 'allPosts') {
            fetchPosts();
        } else if (view === 'myPosts') {
            fetchMyPosts();
        }
    }, [view, selectedFriend]);

    const fetchPosts = () => {
        const token = localStorage.getItem('token');
        $.ajax({
            url: `${API_ROOT}/searchall`,
            method: 'GET',
            headers: {
                'Authorization': `bearer ${token}`,
                'Content-Type': 'application/json',
            },
            success: (response) => {
                setPosts(response || []);  // Set response or fallback to an empty array
            },
            error: (error) => {
                console.error('Get posts failed:', error);
                setPosts([]);  // Set to empty array on error
            },
        });
    };

    const fetchMyPosts = () => {
        const token = localStorage.getItem('token');
        $.ajax({
            url: `${API_ROOT}/searchmy`,
            method: 'GET',
            headers: {
                'Authorization': `bearer ${token}`,
                'Content-Type': 'application/json',
            },
            success: (response) => {
                setPosts(response || []);  // Set response or fallback to an empty array
            },
            error: (error) => {
                console.error('Get posts failed:', error);
                setPosts([]);  // Set to empty array on error
            },
        });
    };

    const handleDelete = (post) => {
        const token = localStorage.getItem('token');
        const formData = new FormData();
        formData.append('username', post.user);
        formData.append('message', post.message);
        formData.append('timestamp', post.timestamp);
        $.ajax({
            url: `${API_ROOT}/deletepost`,
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${token}`,
            },
            contentType: false,
            processData: false,
            data: formData,
            success: () => {
                fetchMyPosts();},
            error: (error) => {
                console.error('Delete post failed:', error);
            }
        });
    };


    return (
        <div>
            {view === 'allPosts' && (
                posts.length > 0 ? (
                    <div style={{ padding: '20px' }}>
                        <Row gutter={[16, 16]}>
                            {posts.map((post, index) => (
                                <Col key={index} xs={24} sm={12} md={8} lg={6}>
                                    <Card
                                        hoverable
                                        cover={
                                            <img alt={post.message} src={post.url} style={{ width: '100%', height: '400px', objectFit: 'cover' }} />
                                        }
                                    >
                                        <Card.Meta
                                            title={post.user}
                                            description={
                                                <>
                                                    <p>{post.message}</p>
                                                    <p>{new Date(post.timestamp).toLocaleString()}</p>
                                                </>
                                            }
                                        />
                                    </Card>
                                </Col>
                            ))}
                        </Row>
                    </div>
                ) : (
                    <div style={{ padding: '20px', textAlign: 'center' }}>
                        <p>No posts available.</p>
                    </div>
                )
            )}

            {view === 'createPost' && <CreatePost onPostCreated={fetchPosts} />}

            {view === 'myPosts' && (
                posts.length > 0 ? (
                    <div style={{ padding: '20px' }}>
                        <h2>My Posts</h2>
                        <Row gutter={[16, 16]}>
                            {posts.map((post, index) => (
                                <Col key={index} xs={24} sm={12} md={8} lg={6}>
                                    <Card
                                        hoverable
                                        cover={
                                            <img alt={post.message} src={post.url} style={{ width: '100%', height: '400px', objectFit: 'cover' }} />
                                        }
                                        actions={[
                                            <Button type="danger" onClick={() => handleDelete(post)}>x</Button>
                                        ]}
                                    >
                                        <Card.Meta
                                            title={post.user}
                                            description={
                                                <>
                                                    <p>{post.message}</p>
                                                    <p>{new Date(post.timestamp).toLocaleString()}</p>
                                                </>
                                            }
                                        />
                                    </Card>
                                </Col>
                            ))}
                        </Row>
                    </div>
                ) : (
                    <div style={{ padding: '20px', textAlign: 'center' }}>
                        <p>No posts available.</p>
                    </div>
                )
            )}
        </div>
    );
};

export default PostsDisplay;
