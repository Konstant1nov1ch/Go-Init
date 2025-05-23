Отлично. Я опишу визуальную структуру интерфейса веб-сервиса на React для генерации шаблонов, учитывая динамические элементы вроде выбора базы данных и отображения файловой структуры.
Я скоро вернусь с описанием, которое можно будет использовать в генеративной модели для создания интерфейса.

# React Template Generator UI Overview

The interface is a React-based web application for configuring and generating backend project templates. It is organized into three main areas: a **left sidebar** for navigation, a **main configuration panel** for input and options, and a **right-side preview** of the project’s file structure. Below is a detailed description of each part of the interface and its components, written in terms familiar to front-end developers.

## Left Sidebar (Navigation)
- **Sidebar Layout:** A vertical navigation bar is fixed on the left side of the application. It provides quick access to template actions.
- **"Gen New Template" Button:** A prominent button (or menu item) labeled "Gen New Template". Clicking this initiates the creation of a new template configuration (clears current form inputs for a fresh start).
- **"Recent Templates" Button:** A second navigation item labeled "Recent Templates". This lets the user view or load recently saved/generated templates for editing or review.
- **Styling:** The sidebar visually distinguishes the active selection. For example, the "Gen New Template" button might be highlighted when the user is starting a new configuration.

## Main Configuration Panel (Center)
The main panel contains a form-like interface where the user provides details and preferences for the backend template. It includes fields for service information, protocol selection, database options, and advanced settings.

- **Service Name Input:** At the top of the main panel, there is a single-line text input field labeled **"Service Name"**. This field spans the full width of the panel, allowing the user to type the name of the service or project. It might use a standard `<input>` or Material-UI `TextField` component, styled to fill the available width.
- **Protocol Selection (Radio Buttons):** Directly below the service name, the interface presents a set of radio buttons for choosing the API/communication protocol. All options are displayed on one line (horizontally aligned) for easy comparison. The choices include:  
  - **gRPC:** Select this for a gRPC-based service template.  
  - **GQL (GraphQL):** Select this for a GraphQL-based service.  
  - **REST:** Select this for a RESTful API service.  
  The radio group is likely labeled (visibly or via accessibility) as "Protocol". Only one of these can be selected at a time. When a user selects a protocol, it may influence the project structure (reflected in the preview on the right) but does not by itself reveal additional input fields.

## Advanced Configuration (Database Options)
The interface includes optional **advanced settings** that are hidden by default and can be revealed via an **"Advanced" toggle**. These settings include database configuration options. When the user wants to configure database support, they click the Advanced section to reveal these options.

- **Advanced Toggle Button:** A button or collapsible panel labeled **"Advanced"** is displayed, accompanied by an arrow icon indicating it can expand or collapse. By default, this section is collapsed (arrow pointing to the right or down). When clicked, the arrow rotates (e.g., down to up) and the advanced options section becomes visible.
- **Database Selection Panel:** Within the advanced options, a subsection for **Database configuration** appears. This may be presented as a labeled group or dropdown panel for selecting a database. Initially, if the user hasn’t engaged advanced options or has not opted for a database, this panel is hidden.
  - **Database Type Radio Buttons:** If the user chooses to include a database (for example, by selecting a database option or toggling a checkbox), a set of radio buttons is revealed, allowing selection of the database type:  
    - **PostgreSQL (psql):** Option for a PostgreSQL database.  
    - **MySQL:** Option for a MySQL database.  
    - **None:** Option to indicate no database integration.  
    These options are typically grouped under a label like "Database". Only one can be selected. If **"None"** is selected (or if the user hasn’t interacted with this section), it implies no database will be used, and thus the database configuration panel can remain hidden or get hidden again.
  - **Conditional Display:** When **"None"** is selected or if the user collapses the advanced section, the database options panel and any related fields (like DDL upload) are hidden from view. This keeps the UI simple when database integration is not needed.
- **DDL Upload Field (Conditional):** If a database option **other than "None"** is selected (meaning the user wants to include a database), an **upload control** for a DDL file becomes active. This could be an icon or button labeled "DDL" (see Action Icons below) that, when clicked, opens a file picker dialog to upload a DDL (Data Definition Language) script. This DDL file would contain the database schema definitions to be included in the generated template. If no database is selected, this upload control remains disabled or hidden, preventing confusion.

## Action Icons (Export/Import Controls)
At the bottom of the main panel (or in a toolbar area beneath the form), the interface provides two icon-based controls for importing a schema and exporting the generated project. These are represented as arrow icons with short labels, functioning like buttons:

- **DDL Import Icon:** An icon depicting a **downward arrow**, labeled "DDL". This control is used for importing a database schema definition (DDL file) into the template generator. It is **active only when a database is selected** (i.e., when the database radio option is set to either PostgreSQL or MySQL). If the user clicks this icon, it triggers a file selection dialog to upload a DDL script from the user’s computer. The downward arrow icon suggests "bringing in" data (importing) into the application. When no database is being used (database type is "None"), this icon is either grayed out (disabled) or not shown at all.
- **ZIP Download Icon:** Next to the DDL icon is an icon showing an **upward arrow**, labeled "ZIP". This represents the action to **download** the generated project as a zip archive. Once the user has configured all options (service name, protocol, possibly database, etc.), they can click this "ZIP" icon to package the generated template and download it. The upward arrow icon signifies "sending out" or exporting data (in this case, exporting the project to the user’s system). This control is generally always available, but it will only produce a meaningful download after the configuration is complete. (In some implementations, clicking this might also trigger the actual generation process on the backend before providing the download.)

*Layout Note:* These two icon buttons are placed together for convenience. They may appear as a small horizontal toolbar or a footer section of the main panel. Each icon has a text label ("ddl" and "zip") for clarity. Tooltips might also be provided on hover (e.g., "Upload Database Schema (DDL)" and "Download Project Zip").

## Project Structure Preview (Right Panel)
On the right side of the interface, there is a dynamic text panel that serves as a **live preview of the project’s file/folder structure**. This helps the user visualize what the generated template will look like as they configure options.

- **Default View:** By default (with minimal configuration), the preview shows a basic project structure. For example, it might display:  
  ```plaintext
  cmd/  
   └── main.go
  ```  
  This indicates that the project will have a directory `cmd` containing a file `main.go` (a typical entry point for Go projects, for instance).
- **Hierarchical Display:** The structure is shown in a hierarchical, indented format (similar to a file explorer tree or an outline). Folders are listed (possibly with folder icons or simply as names with trailing slashes), and files are listed under their respective directories.
- **Dynamic Updates:** As the user selects different options on the main panel, this file structure updates in real time to include additional files and directories that will be generated. Some examples of how the structure might change:
  - If the user **selects a database** (e.g., PostgreSQL or MySQL), new entries related to database configuration might appear. For instance, a `db/` directory could be added to the tree with files for database connection or migrations (e.g., `db/config.go` or migration SQL files). Alternatively, a `migrations/` folder might appear at the top level containing DDL-derived scripts.
  - If the user **chooses gRPC** as the protocol, the structure might include files for service definitions. For example, a `proto/` directory could appear containing `.proto` files (like `service.proto`), and generated code files (stubs) could be listed in appropriate locations (such as a `pb/` or `internal/` package directory for the compiled gRPC code).
  - If **GraphQL (GQL)** is selected, the preview might show files such as `schema.graphql` or a `graphql/` directory with resolver stubs, indicating that GraphQL schema and resolver boilerplate will be included.
  - For **REST**, the structure might include routing or controller files. Perhaps a `api/` or `handlers/` directory with files like `routes.go` or `handler.go` could appear.
  - Enabling **Advanced options** (like the database) together with a specific protocol could combine these additions. The preview will reflect all selected configurations (e.g., a project with both gRPC and a database might show both a `proto/` directory and a `db/` directory in the tree).
- **Usage:** This read-only panel gives immediate feedback. Front-end wise, it could be implemented as a `<pre>`formatted text block or a virtualized list, updated via state when options change. It helps users verify the inclusion of components (like seeing that choosing a database indeed adds the expected files) before downloading the project.
- **Scroll & Size:** If the structure becomes long, the preview panel can scroll. It is sized to occupy the remaining width on the right beside the main panel, ensuring the layout shows both the form and preview side by side.

## Summary of Interaction
1. **Start New Template:** The user clicks "Gen New Template" in the sidebar to begin. The main panel is cleared to default values (empty service name, default protocol selected, advanced options hidden).
2. **Enter Service Name:** The user types a name into the "Service Name" field.
3. **Select Protocol:** The user chooses one of the radio buttons (gRPC, GQL, or REST) for the service’s protocol. The preview panel updates to show files relevant to that choice.
4. **Advanced Options (Database):** If the user wants to include a database, they click the "Advanced" arrow. The database options panel expands. The user selects **PostgreSQL** or **MySQL** (or leaves it as **None** for no database). If a database is selected, the "DDL" upload icon becomes enabled.
5. **Upload DDL (optional):** For database users, the user may click the "DDL" (down-arrow) icon to upload a schema file. (This step is optional if the template can be generated without an explicit schema, but available if they have one.) After uploading, the preview might update further (for example, showing tables or models, depending on how the template generator uses the DDL).
6. **Generate & Download:** Once satisfied with the configuration, the user clicks the "ZIP" (up-arrow) icon. The application back-end generates a new project template with the specified features. A download of a `.zip` file is then triggered, containing the project files and directories as shown in the preview.
7. **Review or Continue:** The user can then unzip and inspect the generated project. If changes are needed, they might adjust options or start a new template again.

## Notes for Implementation (For a Generative Model)
- Each UI element described (sidebar, input field, radio group, buttons, icons, preview tree) can correspond to React components or JSX elements. For example, the sidebar might be a `<nav>` element or a `Drawer` component containing `<Button>` components for each action.
- State management is crucial: selecting protocol or database updates the preview. In React, consider using state hooks or context to handle form inputs and reflect changes in the preview component.
- Conditional rendering is used for advanced options and the DDL upload control. For instance, `{showDatabaseOptions && <DatabaseOptions />}` could toggle the database radio group, and `{hasDatabase && <DdlUploadButton />}` could control the DDL icon’s active state.
- The preview panel could be a component that takes the current configuration as props (service name, protocol, db choice) and calculates the file structure. It might use a tree data structure to render folders and files. Simpler approach: conditionally include sections of text.
- Use clear labels and maybe tooltips for icons to ensure usability (since "DDL" and "ZIP" might not be immediately obvious to all users without context).

By following this description, a generative model or a front-end developer can construct the React components and layout to match the intended interface. Each part of the UI is defined in terms of structure and behavior, which can guide the coding of the components and their styling.