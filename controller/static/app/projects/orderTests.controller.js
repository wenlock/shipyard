(function(){
    'use strict';

    angular
        .module('shipyard.projects')
        .controller('OrderTestsController', OrderTestsController);

    OrderTestsController.$inject = ['orderTests', 'project', '$scope', 'ProjectService', '$stateParams'];
    function OrderTestsController(orderTests, project, $scope, ProjectService, $stateParams) {
        var vm = this;
        vm.projectDetails = project;
        vm.testsToOrder = orderTests;

    }
})();
