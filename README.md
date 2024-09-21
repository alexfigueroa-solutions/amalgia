# **Amalgia**

Amalgia is a CLI (Command-Line Interface) application built with Go and the Bubble Tea framework. It empowers you to:

- Import your resume and cover letter via a terminal-based file explorer (to be implemented).
- Fetch README files from all your GitHub repositories (both public and private).
- Save the fetched README files to a local directory for further processing.
- **Interact with your professional profile using OpenAI's API**, enabling you to generate documents like resumes or cover letters, and even chat with your profile data.
- Lay the groundwork for generating an up-to-date professional profile by combining your existing documents with your GitHub project information.

## **Table of Contents**

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [Project Structure](#project-structure)
- [Configuration](#configuration)
- [Next Steps](#next-steps)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)
- [Contact](#contact)

---

## **Features**

- **GitHub Integration**: Fetches README files from all your GitHub repositories, including private ones.
- **Local Storage**: Saves the README files to a `readmes` directory for easy access and processing.
- **OpenAI API Integration**:
  - **Chat with Your Profile**: Interact with your professional data to gain insights or prepare for interviews.
  - **Document Generation**: Use AI to generate resumes, cover letters, or other professional documents based on your profile and project data.
- **Extensible Framework**: Built using Bubble Tea, allowing for easy expansion and customization of the terminal UI.
- **Placeholder for File Import**: A foundation is set for importing your resume and cover letter via a terminal-based file explorer (implementation pending).

---

## **Prerequisites**

- **Go Environment**: Go 1.16 or later installed on your system.
- **GitHub Personal Access Token**:
  - Required to access your repositories, especially private ones.
  - Generate one from your GitHub account settings with the necessary scopes (`repo` scope is required).
  - [GitHub Token Generation Guide](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token)
- **OpenAI API Key**:
  - Needed to access OpenAI's services for generating documents and interacting with your profile.
  - Sign up and obtain an API key from [OpenAI's website](https://beta.openai.com/signup/).

---

## **Installation**

### **1. Clone the Repository**

```bash
git clone https://github.com/alexfigueroa-solutions/amalgia.git
cd amalgia
```

### **2. Set Up Environment Variables**

#### **GitHub Token**

Set your GitHub Personal Access Token as an environment variable:

```bash
export GITHUB_TOKEN=your_personal_access_token
```

**Note**: Replace `your_personal_access_token` with the token you generated.

#### **OpenAI API Key**

Set your OpenAI API key as an environment variable:

```bash
export OPENAI_API_KEY=your_openai_api_key
```

**Note**: Replace `your_openai_api_key` with the API key you obtained from OpenAI.

### **3. Install Dependencies**

Run the following command to install the required Go modules:

```bash
go mod tidy
```

---

## **Usage**

Run the application using:

```bash
go run main.go
```

### **Application Flow**

1. **Start the Application**: Upon running, the application initializes and starts the Bubble Tea program.
2. **Loading Data**:
   - The application attempts to import your resume and cover letter (currently a placeholder function).
   - It fetches the README files from your GitHub repositories.
3. **OpenAI Interaction**:
   - Use the integrated OpenAI API to generate professional documents or interact with your profile data.
4. **Saving READMEs**:
   - README files are saved to the `readmes` directory within the project.
   - The application lists the repositories whose READMEs were successfully saved.
5. **User Interaction**:
   - Navigate through the terminal UI to access different features.
   - Press `q` or `ctrl+c` to exit the application.

### **Expected Output**

```
Files Imported:
- /path/to/resume.pdf
- /path/to/cover_letter.pdf

Projects Fetched from GitHub and READMEs saved:
- repository1
- repository2
- repository3

AI-Powered Actions:
- Generate Resume
- Generate Cover Letter
- Chat with Your Profile

Press q to quit.
```

---

## **Project Structure**

```
amalgia/
├── main.go          # Main application file
├── go.mod           # Go module file
├── go.sum           # Go checksum file
├── readmes/         # Directory where README files are saved
├── README.md        # This README file
└── config/          # Configuration files and templates
```

---

## **Configuration**

### **Environment Variables**

Ensure the following environment variables are set:

- `GITHUB_TOKEN`: Your GitHub Personal Access Token.
- `OPENAI_API_KEY`: Your OpenAI API Key.

### **Templates**

You can customize document templates located in the `config/templates` directory (to be implemented). These templates will be used by the OpenAI API to generate personalized documents.

---

## **Next Steps**

### **Implement File Explorer Functionality**

- **Objective**: Replace the placeholder `getResumeAndCoverLetter` function with actual file selection logic.
- **Approach**:
  - Utilize Bubble Tea's `filetree` or `filepicker` components.
  - Allow users to navigate their file system within the terminal to select their resume and cover letter.

### **Integrate OpenAI API Functionality**

- **Objective**: Enable AI-powered interactions and document generation.
- **Approach**:
  - Use the OpenAI Go client library or make direct API calls.
  - Implement commands or menu options in the terminal UI for AI features.
  - Handle API responses and errors gracefully.

### **Process README Contents**

- **Objective**: Parse and extract meaningful information from the README files.
- **Approach**:
  - Use libraries like `goldmark` for Markdown parsing.
  - Extract project descriptions, technologies used, and key accomplishments.

### **Generate an Up-to-Date Resume**

- **Objective**: Combine your existing resume, cover letter, and GitHub project information into a cohesive, updated professional profile.
- **Approach**:
  - Use the OpenAI API to generate or enhance your resume based on your data.
  - Incorporate templating for consistent formatting.
  - Allow customization of generated content.

### **Enhance Error Handling and User Feedback**

- Provide informative messages in case of errors.
- Implement logging to track the application's behavior and issues.

### **Improve the User Interface**

- Customize the Bubble Tea interface for a better user experience.
- Add navigation instructions and visual enhancements.
- Implement interactive menus for AI features.

---

## **Contributing**

Contributions are welcome! If you'd like to help improve Amalgia, please follow these steps:

1. **Fork the Repository**

2. **Create a Feature Branch**

   ```bash
   git checkout -b feature/YourFeature
   ```

3. **Commit Your Changes**

   ```bash
   git commit -m "Add your message"
   ```

4. **Push to the Branch**

   ```bash
   git push origin feature/YourFeature
   ```

5. **Open a Pull Request**

---

## **License**

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## **Acknowledgments**

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**: For the excellent TUI framework.
- **[go-github](https://github.com/google/go-github)**: For the GitHub API client library.
- **[OpenAI](https://openai.com/)**: For providing the API to enable AI-powered features.
- **[ChatGPT](https://openai.com/)**: For assistance in generating the initial project setup.

---

## **Contact**

For any questions or suggestions, please contact:

- **Name**: Alex Figueroa
- **Email**: [alex@alexfigueroa.solutions](mailto:alex@alexfigueroa.solutions)
- **GitHub**: [alexfigueroa-solutions](https://github.com/alexfigueroa-solutions)

---

**Happy coding with Amalgia!**
