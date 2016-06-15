(function(){
    'use strict';

    angular
        .module('shipyard.projects')
        .controller('BuildResultsController', BuildResultsController);

    BuildResultsController.$inject = ['buildResults', '$scope', 'ProjectService', '$stateParams'];
    function BuildResultsController(buildResults, $scope, ProjectService, $stateParams) {
        var vm = this;

        vm.results = map(buildResults.data, function (item) {
            if (item.targetArtifact.artifact.imageId === buildResults.chosenImageId) {
                return item;
            }
        });

        function map(arr, callback) {
            var bin = [];
            for (var i = 0; i < arr.length; i++) {
                var item = callback(arr[i]);
                item && bin.push(item);
            }
            return bin;
        }
    }
})();
