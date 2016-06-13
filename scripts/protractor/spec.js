var config = {
    registryName: 'local',
    registryAddress: 'https://myd-vm23183.hpswlabs.adapps.hp.com:5002',
    registryUsername: 'admin',
    registryPassword: 'password',
    projectName: 'Selenium Project',
    projectNameOnEdit: 'Selenium Project Edited',
    testName: 'Pied Piper',
    testTagSuccess: 'success',
    testTagFailure: 'failure',
    testDescription: 'test description',
    testNameEdit: 'Test alpine',
    imageName: 'alpine',
    tag: 'latest',
    imageDescriptionEdit: 'new image description',
    editTag: '2.6'
};
// TODO: change selectors to id/model/repeaters
var sy = {
    usernameInputField: by.model('vm.username'),
    passwordInputField: by.model('vm.password'),
    loginSubmitButton: by.css('.ui.blue.submit.button'),
    registriesButton: by.id('registries-button'),
    addNewRegistry: by.css('.ui.small.green.labeled.icon.button'),
    addRegistryName: by.model('vm.name'),
    addRegistryAddress: by.model('vm.addr'),
    addRegistryUsername: by.model('vm.username'),
    addRegistryPassword: by.model('vm.password'),
    addRegistrySkipTLS: by.css('.ui.checkbox'),
    addRegistryButton: by.css('.ui.green.submit.button'),
    addRegistryList: by.repeater('r in filteredRegistries = (vm.registries | filter:tableFilter)'),
    ilmButton: by.id('ilm-button'),
    logoProjectListView: by.id('logo-project-view'),
    logoProjectEditView: by.id('logo-project-edit-view'),
    logoProjectInspectView: by.id('logo-project-inspect-view'),
    createNewProjectButton: by.id('create-new-project-button'),
    createProjectButton: by.id('create-project-button'),
    createProjectName: by.id('create-project-name'),
    createProjectDescription: by.model('vm.project.description'),
    editProjectHeader: by.id('edit-project-header'),
    editProjectName: by.model('vm.project.name'),
    editProjectDescription: by.model('vm.project.description'),
    saveProjectButton: by.id('edit-project-save-project'),
    editProjectSaveSuccess: by.id('edit-project-save-success'),
    editProjectSaveFailure: by.id('edit-project-save-failure'),
    createNewImageButton: by.id('create-new-image-button'),
    createNewTestButton: by.id('create-new-test-button'),
    createImageModal: by.className('ui fullscreen modal transition create image'),
    createImageHeader: by.id('create-image-header'),
    createImageLocation: by.id('create-image-location'),
    createImageLocationPublicReg: by.id('create-image-location-public-reg'),
    createImageNameSearch: by.id('create-image-name-search'),
    createImageTag: by.id('create-image-tag'),
    createImageDescription: by.model('vm.createImage.description'),
    createImageSave: by.id('create-image-save'),
    createImageList: by.repeater('image in vm.images'),
    editImageHeader: by.id('edit-image-header'),
    editImageApply: by.id('edit-image-apply'),
    deleteImageHeader: by.id('delete-image-header'),
    deleteImageButton: by.id('delete-image-button'),
    createTestModal: by.className('ui fullscreen modal transition create test'),
    createTestHeader: by.id('create-test-header'),
    createTestDisplayName: by.id("create-test-display-name"),
    createTestTagSuccess: by.model("vm.createTest.tagging.onSuccess"),
    createTestTagFailure: by.model("vm.createTest.tagging.onFailure"),
    createTestDescription: by.model("vm.createTest.description"),
    createTestProviderDropdown: by.className('ui search fluid dropdown testProvider'),
    createTestProviderMenuTransitioner: by.className('menu transition visible'),
    createTestImagesToTest: by.css('[placeholder="All images"]'),
    createTestImagesToTestCSS: '[placeholder="All images"]',
    createTestSelectImageToTest: by.css('[data-value="' + config.imageName + ':' + config.tag + '"]'),
    createTestSelectImageToTestCSS: '[data-value="' + config.imageName + ':' + config.tag + '"]',
    createTestSaveButton: by.id('test-create-save-button'),
    editTestHeader: by.id('edit-test-header'),
    editTestApply: by.id('edit-test-apply'),
    editTestDisplayName: by.model('vm.editTest.name'),
    deleteTestHeader: by.id('delete-test-header'),
    deleteTestButton: by.id('delete-test-button'),
    editProjectList: by.repeater('test in vm.tests'),
    editProjectLoadingMsgNegative: by.css('#content > div.ui.padded.grid.ng-scope > div > div > div.ui.icon.message.negative'),
    editProjectLoadingMsgNegativeCSS: '#content > div.ui.padded.grid.ng-scope > div > div > div.ui.icon.message.negative',
    editProjectLoadingMsg: by.css('#content > div.ui.padded.grid.ng-scope > div > div > div:nth-child(1)'),
    editProjectLoadingMsgCSS: '#content > div.ui.padded.grid.ng-scope > div > div > div:nth-child(1)',
    editProjectGoToProjectsButton: by.css('#content > div.ui.padded.grid.ng-scope > div > div > div.ui.segment.page > div > div.column.row > div > h3 > span > a'),
    projectListTableOfProjects: by.repeater('a in filteredProjects = (vm.projects | filter:tableFilter)'),
    buildQueueTable: by.repeater('project in filtered = (vm.projects | filter:tableFilter) track by $index'),
    deleteProjectHeader: by.id('delete-project-header'),
    deleteProjectButton: by.id('delete-project-button'),
    inspectViewBuilds: by.repeater('test in vm.results.testResults'),
    inspectViewTestName: by.id('inspect-view-test-name'),
    inspectViewMagnifyingGlass: by.id('inspect-view-test-results'),
    inspectProjectGoToProjectsButton: by.id('inspect-go-to-projects')
};

describe('ILM', function() {
    it('should have a title', function() {
        console.log("check title");
        browser.get('http://'+process.env.SHIPYARD_HOST);
        expect(browser.getTitle()).toEqual('shipyard');
    });

    it('should be able to login', function() {
        console.log("login into shipyard");
        element(sy.usernameInputField).sendKeys('admin ');
        element(sy.passwordInputField).sendKeys('shipyard');
        element(sy.loginSubmitButton).click();
    });

    it('should be able to navigate to registries tab', function() {
        console.log("navigate to registries view");
        element(sy.registriesButton).click();
        expect(element(sy.addNewRegistry).isDisplayed()).toBeTruthy();
    });

    it('should be able to add a new registry', function() {
        console.log("add new registry");
        element(sy.addNewRegistry).click();
        expect(element(by.css('.ui.dividing.header')).isDisplayed()).toBeTruthy();
        element(sy.addRegistryName).sendKeys(config.registryName);
        element(sy.addRegistryAddress).sendKeys(config.registryAddress);
        element(sy.addRegistryUsername).sendKeys(config.registryUsername);
        element(sy.addRegistryPassword).sendKeys(config.registryPassword);
        element(sy.addRegistrySkipTLS).click();
        element(sy.addRegistryButton).click();
        /*var registryDetails = element(sy.addRegistryList.row(0));
         var registry = registryDetails.all(by.tagName('td'));
         expect(registry.get(0).getText()).toEqual(config.registryName);
         expect(registry.get(1).getText()).toEqual(config.registryAddress);*/
    });

    it('should be able to navigate to project list', function() {
        console.log("navigate to project list view");
        element(sy.ilmButton).click();
    });

    it('should have a logo', function() {
        console.log("check logo");
        expect(element(sy.logoProjectListView).isDisplayed()).toBeTruthy();
    });

    it('should have message that there are no projects', function() {
        console.log("check the presence of 'there are no projects' message");
        expect(element(by.id('empty-project-list-message')).isDisplayed()).toBeTruthy();
    });

    it('should have the refresh button', function() {
        console.log("should have a refresh button in the project list view");
        expect(element(by.id('refresh-project-button')).isDisplayed()).toBeTruthy();
        element(by.id('refresh-project-button')).click();
    });

    it('should be able to navigate to the project create view', function() {
        console.log("navigate to project create view");
        element(sy.createNewProjectButton).click();
    });

    it('should be able to create a new project', function() {
        console.log("create a new project");
        element(sy.createProjectName).sendKeys(config.projectName);
        element(sy.createProjectDescription).sendKeys('Description1');
        element(sy.createProjectButton).click();
    });

    it('project should be successfully created', function() {
        console.log("check if the project was successfully created");
        expect(
            element(sy.editProjectHeader).getText()
        ).toEqual(
            'Project ' + config.projectName
        );

        expect(
            element(sy.editProjectName).getAttribute('value')
        ).toEqual(
            config.projectName
        );

        expect(
            element(sy.editProjectDescription).getAttribute('value')
        ).toEqual(
            'Description1'
        );
    });

    it('should be able to modify an existing project', function() {
        console.log("modify an existing project");
        expect(element(sy.saveProjectButton).getAttribute('class')).toEqual('ui small disabled button');
        element(sy.editProjectName).clear();
        element(sy.editProjectName).sendKeys(config.projectNameOnEdit);
        element(sy.saveProjectButton).click();
        expect(
            element(sy.editProjectName).getAttribute('value')
        ).toEqual(
            config.projectNameOnEdit
        );
    });

    it('should have a logo in the edit project view', function() {
        console.log("check logo in edit project view");
        expect(element(sy.logoProjectEditView).isDisplayed()).toBeTruthy();
    });

    it('should see tooltip for disable image verification when hovering over it', function() {
        console.log("open tooltip when hovering over disable image verification question mark");
        browser.actions().mouseMove(element(by.id('help-circle-image-verification'))).perform();
        browser.sleep(1000);
        expect(element(by.id('image-verification-disable-tooltip')).isDisplayed()).toBeTruthy();
    });

    it('check tooltip message changes after clicking the disable image verification slider', function() {
        console.log("tooltip message changes after clicking the disable image verification slider");
        browser.actions().mouseMove(element(by.id('help-circle-image-verification'))).perform();
        browser.sleep(1000);
        expect(element(by.id('image-verification-disable-tooltip')).isDisplayed()).toBeTruthy();
        expect(element(by.id('image-verification-disable-tooltip')).getText()).toEqual('Click the slider to disable image verification.');
        element(by.id('image-verification-slider')).click();
        browser.actions().mouseMove(element(by.id('help-circle-image-verification'))).perform();
        browser.sleep(1000);
        expect(element(by.id('image-verification-enable-tooltip')).isDisplayed()).toBeTruthy();
        expect(element(by.id('image-verification-enable-tooltip')).getText()).toEqual('Click the slider to enable image verification.');
        element(by.id('image-verification-slider')).click();
    });

    it('should open modal window for add image', function() {
        console.log("open modal window for add image");
        element(sy.createNewImageButton).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.createImageModal)), 60000);
        expect(element(sy.createImageModal).isDisplayed()).toBe(true);
        expect(element(sy.createImageHeader).getText()).toEqual('Add Image');
    });

    it('should add new image from public registry', function() {
        console.log("add new image from public registry");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.createImageLocation), 60000));
        element(sy.createImageLocation).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.createImageLocationPublicReg), 60000));
        element(sy.createImageLocationPublicReg).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.createImageNameSearch), 60000));
        element(sy.createImageNameSearch).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.createImageNameSearch), 60000));
        element(sy.createImageNameSearch).sendKeys(config.imageName);
        browser.wait(protractor.until.elementLocated(by.className('description')), 180000);
        browser.wait(protractor.ExpectedConditions.visibilityOf(element.all(by.className('description')).get(0), 60000));
        element.all(by.className('description')).get(0).click();
        element(sy.createImageTag).click();
        browser.wait(protractor.until.elementLocated(by.id('tag-results')), 180000);
        browser.wait(protractor.ExpectedConditions.visibilityOf(element.all(by.id('tag-results')).get(0), 60000));
        element.all(by.id('tag-results')).get(0).click();
        element(sy.createImageDescription).sendKeys('image description');
        browser.wait(protractor.until.elementLocated(sy.createImageSave), 180000);
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.createImageSave), 60000));
        element(sy.createImageSave).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.createImageList.row(0)), 60000));
        var imageDetails = element(sy.createImageList.row(0));
        var image = imageDetails.all(by.tagName('td'));
        expect(image.get(1).getText()).toEqual(config.imageName);
        expect(image.get(3).getText()).toEqual(config.tag);
    });

    it('should be able to edit an existing image', function() {
        console.log("edit image");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.createImageList.row(0)), 60000));
        var imageDetails = element(sy.createImageList.row(0));
        imageDetails.element(by.css('i[class="pencil icon"]')).click();
        expect(element(sy.editImageHeader).getText()).toEqual(config.imageName);
        //element(by.id('edit-image-tag')).click();
        //browser.wait(protractor.until.elementLocated(by.id('edit-image-tag-results')), 180000);
        //browser.wait(protractor.ExpectedConditions.visibilityOf(element.all(by.id('edit-image-tag-results')).get(1), 180000));
        //element.all(by.id('edit-image-tag-results')).get(1).click();
        //browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.editImageApply), 180000));
        element(by.model('vm.editImage.description')).clear();
        element(by.model('vm.editImage.description')).sendKeys(config.imageDescriptionEdit);
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.editImageApply), 60000));
        element(sy.editImageApply).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.createImageList.row(0)), 60000));
        var imageDetails = element(sy.createImageList.row(0));
        var image = imageDetails.all(by.tagName('td'));
        expect(image.get(1).getText()).toEqual(config.imageName);
        expect(image.get(4).getText()).toEqual(config.imageDescriptionEdit);
    });

    it('should see tooltip for disable test verification when hovering over it', function() {
        console.log("open tooltip when hovering over disable test verification question mark");
        browser.actions().mouseMove(element(by.id('help-circle-test-verification'))).perform();
        browser.sleep(1000);
        expect(element(by.id('test-verification-disable-tooltip')).isDisplayed()).toBeTruthy();
    });

    it('check tooltip message changes after clicking the disable test verification slider', function() {
        console.log("tooltip message changes after clicking the disable test verification slider");
        browser.actions().mouseMove(element(by.id('help-circle-test-verification'))).perform();
        browser.sleep(1000);
        expect(element(by.id('test-verification-disable-tooltip')).isDisplayed()).toBeTruthy();
        expect(element(by.id('test-verification-disable-tooltip')).getText()).toEqual('Click the slider to disable test verification.');
        element(by.id('test-verification-slider')).click();
        browser.actions().mouseMove(element(by.id('help-circle-test-verification'))).perform();
        browser.sleep(1000);
        expect(element(by.id('test-verification-enable-tooltip')).isDisplayed()).toBeTruthy();
        expect(element(by.id('test-verification-enable-tooltip')).getText()).toEqual('Click the slider to enable test verification.');
        element(by.id('test-verification-slider')).click();
    });

    it('should open the tests modal window', function() {
        console.log("open the add test modal window");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.createNewTestButton), 60000));
        //browser.wait(protractor.ExpectedConditions.elementToBeClickable(element(sy.createNewTestButton), 60000));
        element(sy.createNewTestButton).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.createTestModal), 60000));
        expect(element(sy.createTestModal).isDisplayed()).toBe(true);
        //browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.createTestHeader), 60000));
        expect(element(sy.createTestHeader).getText()).toEqual('Add Test');
    });

    it('should add new test that references the image', function() {
        console.log("add new test that references the image");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.createTestProviderDropdown), 60000));
        element(sy.createTestProviderDropdown).click();
        //browser.wait(protractor.until.elementLocated(sy.createTestProviderMenuTransitioner), 60000);
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(by.css('div[data-value="Clair [Internal]"]')), 60000));
        $('div[data-value="Clair [Internal]"]').click();
        browser.wait(protractor.ExpectedConditions.visibilityOf($(sy.createTestImagesToTestCSS)), 60000);
        element(sy.createTestImagesToTest).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf($(sy.createTestSelectImageToTestCSS)), 60000);
        element(sy.createTestSelectImageToTest).click();
        element(sy.createTestDisplayName).sendKeys(config.testName);
        element(sy.createTestTagSuccess).sendKeys(config.testTagSuccess);
        element(sy.createTestTagFailure).sendKeys(config.testTagFailure);
        element(sy.createTestDescription).sendKeys(config.testDescription);
        browser.sleep(2000);
        expect(
            element(sy.createTestDisplayName).getAttribute('value')
        ).toEqual(
            config.testName
        );
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.createTestSaveButton), 60000));
        element(sy.createTestSaveButton).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.editProjectList.row(0)), 60000));
        var testDetails = element(sy.editProjectList.row(0));
        var test = testDetails.all(by.tagName('td'));
        expect(test.get(1).getText()).toEqual(config.testName);
    });

    it('should be able to edit an existing test', function() {
        console.log("edit test");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.editProjectList.row(0)), 60000));
        var testDetails = element(sy.editProjectList.row(0));
        browser.wait(protractor.ExpectedConditions.visibilityOf(testDetails.element(by.css('i[class="pencil icon"]'))), 60000);
        testDetails.element(by.css('i[class="pencil icon"]')).click();
        expect(element(sy.editTestHeader).getText()).toEqual(config.testName);
        element(sy.editTestDisplayName).clear();
        element(sy.editTestDisplayName).sendKeys(config.testNameEdit);
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.editTestApply), 60000));
        element(sy.editTestApply).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.editProjectList.row(0)), 60000));
        var testDetails = element(sy.editProjectList.row(0));
        var test = testDetails.all(by.tagName('td'));
        expect(test.get(1).getText()).toEqual(config.testNameEdit);
    });

    it('should be able to run the test', function() {
        console.log("run the test");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.editProjectList.row(0)), 60000));
        // Click the play icon for the test
        var buildButton = element(sy.editProjectList.row(0));
        browser.wait(protractor.ExpectedConditions.visibilityOf(buildButton.element(by.css('i[class="play icon"]'))), 60000);
        buildButton.element(by.css('i[class="play icon"]')).click();
        // Wait for status messages / test to build
        browser.wait(protractor.ExpectedConditions.visibilityOf($(sy.editProjectLoadingMsgCSS)), 60000);
        // In this case, we expect the test to fail
        browser.wait(protractor.ExpectedConditions.visibilityOf($(sy.editProjectLoadingMsgNegativeCSS)), 600000);
        // Expect the message to have the `negative` class (sice build will fail)
        expect(element(sy.editProjectLoadingMsg).getAttribute('class')).toBe('ui icon message negative');
    });

    it('should be able return to project listing via the `Go To Projects` icon', function() {
        console.log("return to the project listing using the 'Go To Projects' action item");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.editProjectGoToProjectsButton), 60000));
        // Click the `Go To Projects` button
        element(sy.editProjectGoToProjectsButton).click();
        var projectName = element.all(sy.projectListTableOfProjects).get(-1)
            .element(by.css('#project-name'));
        browser.wait(protractor.ExpectedConditions.visibilityOf(projectName, 60000));
    });

    it('should be able to enter the project"s inspect view', function() {
        console.log("enter the project's inspect view");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element.all(sy.projectListTableOfProjects).get(-1), 60000));
        // Click the `inspect` button for the project
        element.all(sy.projectListTableOfProjects).get(-1)
            .element(by.className('search icon')).click();
        // Wait for the inspect view to load and assert success
        var inspectHeader = element(by.css('.ui.header .content'));
        browser.wait(protractor.ExpectedConditions.visibilityOf(inspectHeader, 60000));
        expect(inspectHeader.getText()).toEqual('Project Results');
    });

    it('should have a logo in the inspect project view', function() {
        console.log("check logo in inspect project view");
        expect(element(sy.logoProjectInspectView).isDisplayed()).toBeTruthy();
    });

    it('should have the auto refresh action', function() {
        console.log("check auto refresh action in inspect project view");
        expect(element(by.id('auto-refresh-enabled')).isDisplayed()).toBeTruthy();
        element(by.id('auto-refresh-enabled')).click();
        expect(element(by.id('auto-refresh-disabled')).isDisplayed()).toBeTruthy();
        element(by.id('auto-refresh-disabled')).click();
    });

    it('should be able to check the project history', function() {
        console.log("check project history");
        expect(element(by.id('history-project-inspect-view')).isDisplayed()).toBeTruthy();
        /*element(by.id('history-project-inspect-view')).click();
         expect(element(by.id('history-header-inspect-view')).isDisplayed()).toBeTruthy();*/
    });

    it('should have the build we just ran', function() {
        console.log("check the build that we ran");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element.all(sy.inspectViewBuilds).get(-1), 60000));
        // Click the `inspect` button for the project
        var lastBuild = element.all(sy.inspectViewBuilds).get(-1);
        expect(
            lastBuild.element(sy.inspectViewTestName).getText()
        ).toEqual(
            config.testNameEdit
        );
    });

    it('should be able to go to edit view', function() {
        console.log("skip to edit project view");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(by.id('edit-project')), 60000));
        expect(element(by.id('edit-project')).getText()).toEqual('Edit Project');
        element(by.id('edit-project')).click();
        expect(element(sy.editProjectHeader).getText()).toEqual('Project ' + config.projectNameOnEdit);
    });

    it('should be able to inspect the test we run', function() {
        console.log("inspect the results of the test we run");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.editProjectList.row(0)), 60000));
        var inspectButton = element(sy.editProjectList.row(0));
        browser.wait(protractor.ExpectedConditions.visibilityOf(inspectButton.element(by.css('i[class="search icon"]'))), 60000);
        inspectButton.element(by.css('i[class="search icon"]')).click();
        var inspectHeader = element(by.css('.ui.header .content'));
        browser.wait(protractor.ExpectedConditions.visibilityOf(inspectHeader, 60000));
        expect(inspectHeader.getText()).toEqual('Project Results');
    });

    it('should be able return from the inspect view to project listing via the `Go To Projects` icon', function() {
        console.log("return to the project listing from the inspect view using the 'Go To Projects' action item");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.inspectProjectGoToProjectsButton), 60000));
        // Click the `Go To Projects` button
        element(sy.inspectProjectGoToProjectsButton).click();
        var projectName = element.all(sy.projectListTableOfProjects).get(-1)
            .element(by.css('#project-name'));
        browser.wait(protractor.ExpectedConditions.visibilityOf(projectName, 60000));
    });

    it('should be able to execute test from project level', function() {
        console.log("execute test from project level");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.projectListTableOfProjects.row(0)), 60000));
        var buildProject = element(sy.projectListTableOfProjects.row(0));
        buildProject.element(by.className('wrench icon')).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(buildProject.element(by.css('i[class="black play icon"]'))), 60000);
        buildProject.element(by.css('i[class="black play icon"]')).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.buildQueueTable.row(0)), 60000));
        var buildDetails = element(sy.buildQueueTable.row(0));
        var build = buildDetails.all(by.tagName('td'));
        expect(build.get(2).getText()).toEqual('new');
    });

    it('should be able to get to edit project view clicking on the Edit action item from the project list view', function() {
        console.log("get to edit project view by clicking on the edit action item");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.projectListTableOfProjects.row(0)), 60000));
        var editProject = element(sy.projectListTableOfProjects.row(0));
        editProject.element(by.className('wrench icon')).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(editProject.element(by.css('i[class="black edit icon"]'))), 60000);
        editProject.element(by.css('i[class="black edit icon"]')).click();
        expect(element(sy.editProjectHeader).getText()).toEqual('Project ' + config.projectNameOnEdit);
    });

    it('should be able to delete a image', function() {
        console.log("delete image");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.createImageList.row(0)), 60000));
        var imageDetails = element(sy.createImageList.row(0));
        imageDetails.element(by.css('i[class="trash icon"]')).click();
        expect(element(sy.deleteImageHeader).getText()).toEqual("Delete Image: "+config.imageName);
        element(sy.deleteImageButton).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.editProjectHeader), 60000));
        expect(element(sy.createImageList.row(0)).isPresent()).toBeFalsy();
    });

    it('should be able to delete a test', function() {
        console.log("delete test");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.editProjectList.row(0)), 60000));
        var testDetails = element(sy.editProjectList.row(0));
        testDetails.element(by.css('i[class="trash icon"]')).click();
        expect(element(sy.deleteTestHeader).getText()).toEqual("Delete Test: "+config.testNameEdit);
        element(sy.deleteTestButton).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.editProjectHeader), 60000));
        expect(element(sy.editProjectList.row(0)).isPresent()).toBeFalsy();
    });

    it('should be able to delete a project', function() {
        console.log("delete project");
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.editProjectGoToProjectsButton), 60000));
        element(sy.editProjectGoToProjectsButton).click();
        var deleteProject = element(sy.projectListTableOfProjects.row(0));
        deleteProject.element(by.className('wrench icon')).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(deleteProject.element(by.css('i[class="black trash icon"]'))), 60000);
        deleteProject.element(by.css('i[class="black trash icon"]')).click();
        browser.wait(protractor.ExpectedConditions.visibilityOf(element(sy.deleteProjectHeader), 60000));
        element(sy.deleteProjectButton).click();
        expect(element(sy.editProjectList.row(0)).isPresent()).toBeFalsy();
    });

});