(function(){
    'use strict';

    angular
        .module('shipyard.projects')
        .controller('CreateController', CreateController);

    CreateController.$inject = ['$scope', 'ProjectService', 'RegistryService', '$state', '$http'];
    function CreateController($scope, ProjectService, RegistryService, $state, $http) {
        var vm = this;

        vm.project = {};
        vm.project.author = localStorage.getItem('X-Access-Token').split(":")[0];

        vm.saveProject = saveProject;

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
