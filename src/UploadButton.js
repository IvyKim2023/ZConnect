import React from 'react';
import { UploadOutlined } from '@ant-design/icons';
import { Button, Upload } from 'antd';

const UploadButton = ({ onFileSelect }) => {
    const props = {
        beforeUpload: (file) => {
            onFileSelect(file);
            return false; // Prevent automatic upload
        },
        showUploadList: true, // Hide the default upload list
    };

    return (
        <Upload {...props}>
            <Button icon={<UploadOutlined />}>Upload</Button>
        </Upload>
    );
};

export default UploadButton;
