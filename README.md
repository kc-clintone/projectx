# EduPulse AI (Go Version)

EduPulse AI is an intelligent educational platform designed to facilitate academic and professional learning. Powered by Google's Gemini AI, it dynamically generates quizzes based on user-defined topics and analyzes performance to provide personalized feedback and study plans.

## Features

*   **AI-Powered Quiz Generation**: Create quizzes on any topic, tailored to specific grade levels (Elementary to Professional).
*   **Flexible Question Formats**: Support for Multiple Choice (MCQ), Open-Ended, or Mixed question types.
*   **Study Material Integration**: Upload images (notes, diagrams) to generate questions based on specific content.
*   **Adaptive Assessment Modes**:
    *   *Per Question*: Timed strictly per question.
    *   *Full Session*: Manage time across the entire quiz.
*   **Comprehensive Analysis**:
    *   Instant grading and feedback.
    *   Identification of strengths and weaknesses.
    *   Personalized study recommendations.
*   **Session Management**: Tracks user history and progress (stored locally in `sessions.json`).

## Tech Stack

*   **Backend**: Go (Golang)
*   **Frontend**: Vanilla JavaScript, HTML5, Tailwind CSS
*   **AI Engine**: Google Gemini API (`gemini-3-flash-preview`)
*   **Data Storage**: JSON-based flat file storage

## Prerequisites

*   Go (1.21 or later recommended)
*   A Google Cloud Project with the Gemini API enabled
*   An API Key for Gemini

## Setup & Installation

1.  **Navigate to the project directory**:
    ```bash
    cd "/home/tomlee/Desktop/dev/Go version"
    ```

2.  **Install Dependencies**:
    Ensure your `go.mod` is set up, then run:
    ```bash
    go mod tidy
    ```

3.  **Environment Configuration**:
    Create a `.env` file in the root directory and add your Gemini API key:
    ```env
    GEMINI_API_KEY=your_api_key_here
    PORT=8080
    ```

4.  **Run the Application**:
    ```bash
    go run main.go
    ```

5.  **Access the App**:
    Open your browser and navigate to `http://localhost:8080`.

## Usage Guide

1.  **Profile Setup**: Enter your name and academic grade level to calibrate the AI's difficulty.
2.  **Create Quiz**:
    *   Enter a topic (e.g., "Photosynthesis", "Linear Algebra").
    *   Select question type (MCQ, Open, Mixed).
    *   (Optional) Upload an image of study notes.
    *   Choose a timing strategy.
3.  **Take Assessment**: Answer questions within the time limit.
4.  **Review Results**: Receive a detailed breakdown of your performance, including AI-generated feedback on open-ended answers and a targeted study plan.

## Project Structure

*   `main.go`: The core Go server handling API routes, Gemini integration, and session management.
*   `static/`: Contains frontend assets (HTML, JS, CSS).
    *   `js/app.js`: Client-side logic for state management and UI rendering.
*   `sessions.json`: Persisted storage for user sessions and quiz history.