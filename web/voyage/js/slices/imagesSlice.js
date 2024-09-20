import * as reduxjstoolkit from 'https://esm.run/@reduxjs/toolkit';
const { createSlice } = reduxjstoolkit;

const initialState = {
    images: [
        { url: "https://example.com/lion.jpg", thumbnail: "", alt: "Lion" },
        { url: "https://example.com/jungle.jpg", thumbnail: "", alt: "Jungle" }
    ]
};

const imagesSlice = createSlice({
    name: 'images',
    initialState,
    reducers: {
        addImage: (state, action) => {
            state.images.unshift(action.payload);
        },
        deleteImage: (state, action) => {
            state.images.splice(action.payload, 1);
        },
        replaceImages: (state, action) => {
            return action.payload;
        }
    }
});

export const { addImage, deleteImage, replaceImages } = imagesSlice.actions;

export default imagesSlice.reducer;