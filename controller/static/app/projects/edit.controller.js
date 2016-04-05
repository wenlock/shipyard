(function(){
    'use strict';

    angular
        .module('shipyard.projects')
        .controller('EditController', EditController);

    EditController.$inject = ['resolvedProject', '$scope', 'ProjectService', 'RegistryService', '$state'];
    function EditController(resolvedProject, $scope, ProjectService, RegistryService, $state) {
        var vm = this;

        vm.project = resolvedProject;

        /*issues with ngmocksE2E. Commenting for now.*/
        vm.project.author = "admin";
        //vm.project.author = localStorage.getItem('X-Access-Token').split(":")[0];

        //TODO: Should the default state be on or off (i.e. checked or not checked)?
        vm.skipImages = false;
        vm.skipTests = false;

        vm.needsBuild = false;

        vm.selected = {};
        vm.selectedItemCount = 0;

        // Create modal, edit modal namespaces
        vm.createImage = {};
        vm.editImage = {};

        vm.createTest = {};
        vm.createTest.tagging = {};
        vm.createTest.provider = {};
        vm.editTest = {};
        vm.editTest.tagging = {};
        vm.editTest.provider = {};

        vm.createImage.additionalTags = [];

        vm.registries = [];
        vm.images = [];
        vm.publicRegistryTags = [];

        vm.providers = [];
        vm.providerTests = [];

        vm.buttonStyle = "disabled";

        vm.createSaveImage = createSaveImage;
        vm.editSaveImage   = editSaveImage;

        vm.createSaveTest = createSaveTest;

        vm.imageList = imageList;
        vm.showImageEditDialog = showImageEditDialog;
        vm.showImageCreateDialog = showImageCreateDialog;
        vm.showTestEditDialog = showTestEditDialog;
        vm.showDeleteImageDialog = showDeleteImageDialog;
        vm.deleteImage = deleteImage;

        vm.updateProject = updateProject;
        vm.resetValues = resetValues;
        vm.getRegistries = getRegistries;
        vm.checkImage = checkImage;
        vm.checkImagePublicRepository = checkImagePublicRepository;
        vm.getImages = getImages;
        vm.showTestCreateDialog = showTestCreateDialog;
        vm.getTestsProviders = getTestsProviders;
        vm.getJobs = getJobs;
        vm.checkProviderTest = checkProviderTest;
        vm.resetTestValues = resetTestValues;
        vm.showDeleteTestDialog = showDeleteTestDialog;
        vm.deleteTest = deleteTest;
        vm.editSaveTest = editSaveTest;
        vm.setTargets = setTargets;

        vm.getRegistries();

        $scope.$on('ngRepeatFinished', function() {
            $('.ui.sortable.celled.table').tablesort();
        });

        $scope.$watch(function() {
            var count = 0;
            angular.forEach(vm.selected, function (s) {
                if(s.Selected) {
                    count += 1;
                }
            });
            vm.selectedItemCount = count;
        });

        // Remove selected items that are no longer visible
        $scope.$watchCollection('imageList()', function () {
            angular.forEach(vm.selected, function (s) {
                if(vm.selected[s.Id].Selected == true) {
                    var isVisible = false
                    angular.forEach($scope.imageList(), function(c) {
                        if(c.Id == s.Id) {
                            isVisible = true;
                            return;
                        }
                    });
                    vm.selected[s.Id].Selected = isVisible;
                }
            });
            return;
        });

        function imageList() {
            return Object.keys(vm.project.image_list)
        }

        $(".ui.search.fluid.dropdown.registry")
            .dropdown({
                onChange: function(value, text, $selectedItem) {
                    $('#edit-project-image-create-modal').find("input").val("");
                    $('.ui.search.fluid.dropdown.image').dropdown('restore defaults');
                    $('.ui.search.fluid.dropdown.tag').dropdown('restore defaults');
                    vm.createImage.name = "";
                    vm.createImage.tag = "";
                    vm.createImage.description = "";
                    vm.buttonStyle = "disabled";
                    vm.editImage.name = "";
                    vm.editImage.tag = "";
                    vm.editImage.description = "";
                    getImages(text);
                }
            });
        $(".ui.search.fluid.dropdown.tag")
            .dropdown({
                onChange: function(value, text, $selectedItem) {
                    // Search for the image layer of the chosen tag
                    // TODO: find a way to save the layer when the user clicks on the tag name
                    var tagObject = $.grep(vm.publicRegistryTags, function (tag) {
                        return tag.name == value;
                    })[0];

                    if (tagObject)
                        vm.createImage.tagLayer = tagObject.hasOwnProperty('layer')? tagObject.layer : '';

                    $scope.$apply();
                }
            });
        $(".ui.search.fluid.dropdown.edit.tag")
            .dropdown({
                onChange: function(value, text, $selectedItem) {
                    // Search for the image layer of the chosen tag
                    // TODO: find a way to save the layer when the user clicks on the tag name
                    var tagObject = $.grep(vm.publicRegistryTags, function (tag) {
                        return tag.name == value;
                    })[0];

                    if (tagObject)
                        vm.editImage.tagLayer = tagObject.hasOwnProperty('layer')? tagObject.layer : '';

                    $scope.$apply();
                }
            });
        $('.ui.search').search({
            apiSettings: {
                url: '/api/v1/search?q={query}',
                beforeXHR: function(xhr) {
                    xhr.setRequestHeader('X-Access-Token', localStorage.getItem('X-Access-Token'))
                },
                // Little hack to get title to show up (some of the docs don't apply to our version of semantic)
                successTest: function(response) {
                    $.each(response.results, function(index,item) {
                        response.results[index].title = response.results[index].name;
                    });
                    return true;
                }
            },
            onSelect: function(result,response) {
                vm.createImage.name = result.title;
                vm.createImage.description = result.description;
                vm.createImage.tag = "";
                vm.buttonStyle = "disabled";
                $('.ui.search.fluid.dropdown.tag').dropdown('restore defaults');
                ProjectService.getPublicRegistryTags(result.name)
                    .then(function(data) {
                        vm.publicRegistryTags = data;
                    }, function(data) {
                        vm.error = data;
                    });
            },
            minCharacters: 3
        });
        $('.ui.search.edit').search({
            apiSettings: {
                url: '/api/v1/search?q={query}',
                beforeXHR: function(xhr) {
                    xhr.setRequestHeader('X-Access-Token', localStorage.getItem('X-Access-Token'))
                },
                // Little hack to get title to show up (some of the docs don't apply to our version of semantic)
                successTest: function(response) {
                    $.each(response.results, function(index,item) {
                        response.results[index].title = response.results[index].name;
                    });
                    return true;
                }
            },
            onSelect: function(result,response) {
                vm.editImage.name = result.title;
                vm.editImage.description = result.description;
                vm.editImage.tag = "";
                vm.buttonStyle = "disabled";
                vm.publicRegistryTags = "";
                $('.ui.search.fluid.dropdown.tag').dropdown('restore defaults');
                ProjectService.getPublicRegistryTags(result.name)
                    .then(function(data) {
                        vm.publicRegistryTags = data;
                    }, function(data) {
                        vm.error = data;
                    });
            },
            minCharacters: 3
        });

        $(".ui.search.fluid.dropdown.testProvider")
            .dropdown({
                onChange: function(value, text, $selectedItem) {
                    $('#edit-project-test-create-modal').find("input").val("");
                    if(value === "Predefined Provider")
                        getTestsProviders();
                }
            });

        vm.myConfig = {
            create: true,
            valueField: 'item',
            labelField: 'item',
            delimiter: '|',
            placeholder: 'All images',
            onInitialize: function(selectize){
                // receives the selectize object as an argument
            },
            // maxItems: 1
        };

        function showTestCreateDialog() {
            vm.createTest = {};
            vm.createTest.tagging = {};
            vm.createTest.provider = {};
            vm.imagesSelectize = [];
            angular.forEach(vm.project.images, function (image) {
                vm.imagesSelectize.push(image.name);
            });
            vm.items = vm.imagesSelectize.map(function(x) { return {item: x};});
            $('#edit-project-test-create-modal')
                .modal({
                    onHidden: function() {
                        vm.buttonStyle = "disabled";
                        $('#test-create-modal').find("input").val("");
                        $('.ui.dropdown').dropdown('restore defaults');
                        vm.createTest.provider.type ="";
                    }
                })
                .modal('show');
        }

        function setTargets(data) {
            vm.createTest.targets=[];
            angular.forEach(data, function (target) {
                angular.forEach(vm.project.images, function (image) {
                    if(image.name === target) {
                        vm.createTest.targets.push({id: image.id,type: target});
                    }
                });
            });
        }

        function showTestEditDialog(test) {
            vm.editTest = $.extend(true, {}, test);
            vm.selectedEditTest = test;
            vm.buttonStyle = "positive";
            if(test.provider.type === "Predefined Provider") {
                vm.providers = [];
                ProjectService.getProviders()
                    .then(function(data) {
                        console.log(data);
                        vm.providers = data;
                        getJobs(test.provider.name);
                    }, function(data) {
                        vm.error = data;
                    })
            }
            if(test.provider.type === "Clair [Internal]") {
                vm.imagesSelectize = [];
                angular.forEach(vm.project.images, function (image) {
                    vm.imagesSelectize.push(image.name);
                });
                vm.items = vm.imagesSelectize.map(function(x) { return {item: x};});
            }
            console.log(test);
            $('#edit-project-test-edit-modal-'+vm.project.id)
                .modal({
                    closable: false
                })
                .modal('show');
        }

        function resetValues() {
            vm.createImage.name = "";
            vm.createImage.tag = "";
            vm.createImage.description = "";
            vm.buttonStyle = "disabled";
            $('#edit-project-image-create-modal').find("input").val("");
            $('.ui.search.fluid.dropdown.registry').dropdown('restore defaults');
            $('.ui.search.fluid.dropdown.image').dropdown('restore defaults');
            $('.ui.search.fluid.dropdown.tag').dropdown('restore defaults');
        }

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

        function showImageCreateDialog() {
            vm.createImage = {};
            $('#edit-project-image-create-modal')
                .modal({
                    onHidden: function() {
                        $('#edit-project-image-create-modal').find("input").val("");
                        $('.ui.dropdown').dropdown('restore defaults');
                        vm.createImage.location = "";
                    },
                    closable: false
                })
                .modal('show');
        }

        function showImageEditDialog(image) {
            vm.editImage = $.extend(true, {}, image);
            vm.selectedEditImage = image;
            vm.buttonStyle = "positive";
            if(image.location === "Public Registry") {
                ProjectService.getPublicRegistryTags(image.name)
                    .then(function(data) {
                        var tagObject = $.grep(data, function (tag) {
                            return tag.name == image.tag;
                        })[0];

                        if (tagObject && tagObject.hasOwnProperty('layer')) {
                            vm.editImage.tagLayer = tagObject.layer;
                            vm.publicRegistryTags = data;
                        }
                    }, function(data) {
                        vm.error = data;
                    });
            }
            if(image.location === "Shipyard Registry") {
                getImages(image.registry);
            }
            $('#edit-project-image-edit-modal-'+vm.project.id)
                .remove()
                .modal({
                    closable: false
                })
                .modal('show');
        }

        function showDeleteImageDialog(image) {
            vm.selectedImage = image;
            $('#edit-project-delete-image-modal').modal('show');
        }

        function showDeleteTestDialog(test) {
            vm.selectedTest = test;
            $('#edit-project-delete-test-modal').modal('show');
        }

        function deleteImage(image) {
            console.log("delete image " + image.id + " from project " + vm.project.id);
            ProjectService.delete(vm.project.id,image.id)
                .then(function(data) {
                    vm.project.images.splice(vm.project.images.indexOf(image), 1);
                }, function (data) {
                    vm.error = data;
                });
        }

        function deleteTest(test) {
            console.log("delete test " + test.id + " from project " + vm.project.id);
            ProjectService.deleteTest(vm.project.id,test.id)
                .then(function(data) {
                    vm.project.tests.splice(vm.project.tests.indexOf(test), 1);
                }, function (data) {
                    vm.error = data;
                });
        }

        function getRegistries() {
            console.log("get regs");
            vm.registries = [];
            RegistryService.list()
                .then(function(data) {
                    console.log(data);
                    vm.registries = data;
                }, function(data) {
                    vm.error = data;
                })

        }

        function getImages(registry) {
            console.log("get images");
            vm.images = [];
            RegistryService.listRepositories(registry)
                .then(function(data) {
                    console.log(data);
                    vm.images = data;
                }, function(data) {
                    vm.error = data;
                })
        }

        function getTestsProviders() {
            console.log("get providers");
            vm.providers = [];
            ProjectService.getProviders()
                .then(function(data) {
                    console.log(data);
                    vm.providers = data;
                }, function(data) {
                    vm.error = data;
                })
        }

        function getJobs(testProvider) {
            vm.providerTests = [];
            $('.ui.search.fluid.dropdown.providerTest').dropdown('restore defaults');
            angular.forEach(vm.providers, function(provider) {
                if(provider.name === testProvider) {
                    vm.providerTests = provider.providerJobs;
                }
            });
        }

        function checkImage(imageData) {
            vm.buttonStyle = "disabled";
            console.log(" check image " + imageData.name + " with tag " + imageData.tag);
            angular.forEach(vm.images, function (image) {
                if(image.name === imageData.name && image.tag === imageData.tag) {
                    vm.buttonStyle = "positive";
                }
            });
        }

        function checkImagePublicRepository(imageData) {
            vm.buttonStyle = "disabled";
            console.log(" check image " + imageData.name + " with tag " + imageData.tag);
            angular.forEach(vm.publicRegistryTags, function (tag) {
                if(tag.name === imageData.tag) {
                    vm.buttonStyle = "positive";
                }
            });
        }

        function checkProviderTest(providerName,providerTest) {
            vm.buttonStyle = "disabled";
            console.log(" check provider name " + providerName + " with tag " + providerTest);
            angular.forEach(vm.providers, function (provider) {
                if(provider.name === providerName) {
                    angular.forEach(provider.providerJobs, function (test) {
                        if(test.name === providerTest) {
                            vm.buttonStyle = "positive";
                        }
                    });
                }
            });
        }

        function createSaveImage(image) {
            if (vm.project.images == null) {
                vm.project.images = [];
            }
            vm.project.images.push($.extend(true,{},image));
        }

        function createSaveTest(test) {
            if (vm.project.tests == null) {
                vm.project.tests = [];
            }
            vm.project.tests.push($.extend(true,{},test));
        }

        function editSaveImage() {
            vm.selectedEditImage.location = vm.editImage.location;
            vm.selectedEditImage.name = vm.editImage.name;
            vm.selectedEditImage.registry = vm.selectedEditImage.registry;
            vm.selectedEditImage.tag = vm.editImage.tag;
            vm.selectedEditImage.description = vm.editImage.description;
            console.log(vm.selectedEditImage);
        }

        function editSaveTest() {
            vm.selectedEditTest.provider.type = vm.editTest.provider.type;
            vm.selectedEditTest.provider.name = vm.editTest.provider.name;
            vm.selectedEditTest.provider.test = vm.editTest.provider.test;
            vm.selectedEditTest.name = vm.editTest.name;
            vm.selectedEditTest.fromTag = vm.editTest.fromTag;
            vm.selectedEditTest.description = vm.editTest.description;
            vm.selectedEditTest.tagging.onFailure = vm.editTest.tagging.onFailure;
            vm.selectedEditTest.tagging.onSuccess = vm.editTest.tagging.onSuccess;
            vm.selectedEditTest.blocker = vm.editTest.blocker;
            vm.selectedEditTest.targets = vm.editTest.targets;
        }

        function updateProject(project) {
            console.log("update project" + project);
            ProjectService.update(project.id, project)
                .then(function(data) {
                    $state.transitionTo('dashboard.projects');
                }, function(data) {
                    vm.error = data;
                });
        }
    }
})();