(function(){
    'use strict';

    angular
        .module('shipyard.projects')
        .controller('CreateController', CreateController);

    CreateController.$inject = ['$scope', 'ProjectService', '$state', '$http'];
    function CreateController($scope, ProjectService, $state, $http) {
        var vm = this;

        vm.project = {};
        vm.project.images = [];
        vm.project.author = localStorage.getItem('X-Access-Token').split(":")[0];
        vm.project.tests = [];

        // Create modal, edit modal namespaces
        vm.createImage = {};
        vm.editImage = {};

        vm.createTest = {};
        vm.createTest.tagging = [];
        vm.createTest.targets = [];
        vm.createTest.provider = {};
        vm.editTest = {};
        vm.editTest.tagging = {};
        vm.editTest.provider = {};

        vm.createImage.additionalTags = [];

        vm.skipImages = true;
        vm.skipTests= true;

        vm.registries = [];
        vm.images = [];
        vm.publicRegistryTags = [];
        vm.tests = [];
        vm.imagesSelectize= [];
        vm.providers = [];
        vm.providerTests = [];

        vm.parameters = [];

        vm.saveProject = saveProject;
        vm.resetTestValues = resetTestValues;

        vm.buttonStyle = "disabled";

        vm.code = "";
        var possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

        for( var i=0; i < 16; i++ )
            vm.code += possible.charAt(Math.floor(Math.random() * possible.length));
        
        function resetTestValues() {
            vm.createTest.provider.name = "";
            vm.createTest.provider.test = "";
            vm.createTest.targets = "";
            vm.createTest.blocker = "";
            vm.createTest.name = "";
            vm.createTest.fromTag = "";
            vm.createTest.description = "";
            vm.createTest.tagging.onSuccess = "";
            vm.createTest.tagging.onFailure = "";
            vm.buttonStyle = "disabled";
            if(vm.createTest.provider.type === "Clair [Internal]") {
                vm.buttonStyle = "positive";
            }
            $('#test-create-modal').find("input").val("");
            $('.ui.search.fluid.dropdown.providerName').dropdown('restore defaults');
            $('.ui.search.fluid.dropdown.providerTest').dropdown('restore defaults');
        }

        function saveProject(project){
            console.log("saving project");
            console.log(project);
            ProjectService.create(project)
                .then(function(data) {
                    $state.transitionTo('dashboard.edit_project', {id: data.id});
                }, function(data) {
                    vm.error = data;
                });
        }
    }
})();
